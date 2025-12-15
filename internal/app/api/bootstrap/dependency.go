package bootstrap

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/auth"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/metrics"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/shared/logger"
	sharedMetrics "eric-cw-hsu.github.io/scalable-auction-system/internal/shared/metrics"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"

	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	kafkaproducer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/producer"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/producer"

	pgProduct "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/product"
	pgStock "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/stock"
	pgUser "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/user"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres/wallet"

	orderUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/order"
	productUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/product"
	stockUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/stock"
	userUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/user"
	walletUsecase "eric-cw-hsu.github.io/scalable-auction-system/internal/usecase/wallet"
)

type Dependencies struct {
	Router           *gin.Engine
	OrderHandler     *handler.OrderHandler
	ProductHandler   *handler.ProductHandler
	UserHandler      *handler.UserHandler
	StockHandler     *handler.StockHandler
	WalletHandler    *handler.WalletHandler
	SyncService      *SyncService
	TokenService     user.TokenService
	MetricsCollector sharedMetrics.MetricsCollector

	// base infra
	logger              *logger.Logger
	pg                  *postgres.PostgresClient
	redis               *redis.RedisClient
	orderReservedWriter *kafka.Writer
}

func InitDependencies(ctx context.Context) (*Dependencies, error) {
	// config
	pgConfig := config.LoadPostgresConfig()
	jwtConfig := config.LoadJWTConfig()
	redisConfig := config.LoadRedisConfig()
	kafkaConfig := config.LoadKafkaConfig()

	// Create and set the global logger
	kafkaSender := logger.NewKafkaSender(kafkaConfig.Brokers, "service.logs")
	logger.AddSender(kafkaSender)

	// Use global logger functions
	logger.Info("Starting API service", map[string]interface{}{
		"kafka_brokers": kafkaConfig.Brokers,
	})

	// Redis
	redisClient := redis.NewRedisClient(redisConfig)
	redisStockCache, err := stockcache.NewRedisStockCache(redisClient)
	if err != nil {
		return nil, fmt.Errorf("inventory repo: %w", err)
	}

	// Postgres
	pgClient, err := postgres.NewPostgresClient(pgConfig)
	if err != nil {
		return nil, fmt.Errorf("pg connect: %w", err)
	}
	pgStockRepo := pgStock.NewPostgresStockRepository(pgClient)
	pgProdRepo := pgProduct.NewPostgresProductRepository(pgClient)
	pgUserRepo := pgUser.NewPostgresUserRepository(pgClient)
	pgWalletRepo := wallet.NewPostgresWalletRepository(pgClient)

	// Kafka
	orderReservedWriter := kafkaInfra.NewWriter(kafkaConfig.Brokers, "order.reserved")
	orderReservedKafkaProducer := kafkaproducer.NewKafkaProducer(orderReservedWriter)

	// Services
	tokenService := auth.NewJWTService(jwtConfig)
	userService := user.NewUserService()

	// Shared Service
	walletProducer := producer.NewEmptyProducer()
	walletUsecaseService := walletUsecase.NewWalletService(pgWalletRepo, walletProducer)

	// Usecases
	placeOrderUsecase := orderUsecase.NewPlaceOrderUsecase(orderReservedKafkaProducer, redisStockCache, walletUsecaseService)
	createProductUsecase := productUsecase.NewCreateProductUsecase(pgProdRepo)
	loginUserUsecase := userUsecase.NewLoginUserUsecase(pgUserRepo, userService, tokenService)
	registerUserUsecase := userUsecase.NewRegisterUserUsecase(pgUserRepo, userService, walletUsecaseService)
	putOnMarketUsecase := stockUsecase.NewPutOnMarketUsecase(pgStockRepo, redisStockCache)
	addFundUsecase := walletUsecase.NewAddFundUsecase(pgWalletRepo, walletProducer)
	subtractFundUsecase := walletUsecase.NewSubtractFundUsecase(pgWalletRepo, walletProducer)
	createWalletUsecase := walletUsecase.NewCreateWalletUsecase(pgWalletRepo, walletProducer)

	// Handlers
	orderHandler := handler.NewOrderHandler(placeOrderUsecase)
	productHandler := handler.NewProductHandler(createProductUsecase)
	userHandler := handler.NewUserHandler(loginUserUsecase, registerUserUsecase)
	stockHandler := handler.NewStockHandler(putOnMarketUsecase)
	walletHandler := handler.NewWalletHandler(addFundUsecase, subtractFundUsecase, createWalletUsecase)

	// Sync Service
	sync := NewSyncService(pgStockRepo, redisStockCache)
	if err := sync.SyncStockToRedis(ctx); err != nil {
		return nil, fmt.Errorf("sync stock: %w", err)
	}

	// Metrics
	metricsCollector := metrics.NewPrometheusCollector()

	return &Dependencies{
		OrderHandler:     orderHandler,
		ProductHandler:   productHandler,
		UserHandler:      userHandler,
		StockHandler:     stockHandler,
		WalletHandler:    walletHandler,
		SyncService:      sync,
		TokenService:     tokenService,
		MetricsCollector: metricsCollector,
	}, nil
}

func (d *Dependencies) Shutdown() {
	logger.Info("Shutting down API service")

	if d.pg != nil {
		if err := d.pg.Close(); err != nil {
			logger.Error("Error closing Postgres connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if d.redis != nil {
		if err := d.redis.Close(); err != nil {
			logger.Error("Error closing Redis connection", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if d.orderReservedWriter != nil {
		if err := d.orderReservedWriter.Close(); err != nil {
			logger.Error("Error closing Kafka writer", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	if err := d.logger.Close(); err != nil {
		logger.Error("Error closing logger", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("API service shut down completed")
}
