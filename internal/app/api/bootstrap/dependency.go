package bootstrap

import (
	"context"
	"fmt"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/config"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/user"
	walletDomain "eric-cw-hsu.github.io/scalable-auction-system/internal/domain/wallet"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/auth"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/postgres"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/redis/stockcache"
	"github.com/gin-gonic/gin"

	kafkaInfra "eric-cw-hsu.github.io/scalable-auction-system/internal/infrastructure/kafka"
	"eric-cw-hsu.github.io/scalable-auction-system/internal/interface/http/handler"
	kafkaproducer "eric-cw-hsu.github.io/scalable-auction-system/internal/interface/kafka/producer"

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
	Router         *gin.Engine
	OrderHandler   *handler.OrderHandler
	ProductHandler *handler.ProductHandler
	UserHandler    *handler.UserHandler
	StockHandler   *handler.StockHandler
	WalletHandler  *handler.WalletHandler
	SyncService    *SyncService
	TokenService   user.TokenService
}

func InitDependencies(ctx context.Context) (*Dependencies, error) {
	// config
	pgConfig := config.LoadPostgresConfig()
	jwtConfig := config.LoadJWTConfig()
	redisConfig := config.LoadRedisConfig()
	kafkaConfig := config.LoadKafkaConfig()

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
	kWriter := kafkaInfra.NewWriter(kafkaConfig.Brokers, "order.placed")
	kafkaProducer := kafkaproducer.NewOrderProducer(kWriter)

	// Services
	tokenService := auth.NewJWTService(jwtConfig)
	userService := user.NewUserService()

	// Event Publisher
	eventPublisher := walletDomain.NewNoOpEventPublisher()

	// Shared Service
	walletUsecaseService := walletUsecase.NewWalletService(pgWalletRepo, eventPublisher)

	// Usecases
	placeOrderUsecase := orderUsecase.NewPlaceOrderUsecase(kafkaProducer, redisStockCache, walletUsecaseService)
	createProductUsecase := productUsecase.NewCreateProductUsecase(pgProdRepo)
	loginUserUsecase := userUsecase.NewLoginUserUsecase(pgUserRepo, userService, tokenService)
	registerUserUsecase := userUsecase.NewRegisterUserUsecase(pgUserRepo, userService, walletUsecaseService)
	putOnMarketUsecase := stockUsecase.NewPutOnMarketUsecase(pgStockRepo, redisStockCache)
	addFundUsecase := walletUsecase.NewAddFundUsecase(pgWalletRepo, eventPublisher)
	subtractFundUsecase := walletUsecase.NewSubtractFundUsecase(pgWalletRepo, eventPublisher)
	createWalletUsecase := walletUsecase.NewCreateWalletUsecase(pgWalletRepo, eventPublisher)

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

	return &Dependencies{
		OrderHandler:   orderHandler,
		ProductHandler: productHandler,
		UserHandler:    userHandler,
		StockHandler:   stockHandler,
		WalletHandler:  walletHandler,
		SyncService:    sync,
		TokenService:   tokenService,
	}, nil
}
