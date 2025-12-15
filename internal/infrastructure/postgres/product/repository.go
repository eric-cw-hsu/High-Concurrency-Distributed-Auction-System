package product

import (
	"context"
	"database/sql"

	"eric-cw-hsu.github.io/scalable-auction-system/internal/domain/product"
)

type ProductRepository struct {
	pgClient *sql.DB
}

func NewPostgresProductRepository(pgClient *sql.DB) *ProductRepository {
	return &ProductRepository{pgClient: pgClient}
}

func (r *ProductRepository) GetProductByID(ctx context.Context, productID string) (*product.Product, error) {
	query := `SELECT id, name, description FROM products WHERE id = $1`
	row := r.pgClient.QueryRow(query, productID)

	var product product.Product
	if err := row.Scan(&product.ID, &product.Name, &product.Description); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Product not found
		}
		return nil, err // Other error
	}

	return &product, nil
}

func (r *ProductRepository) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	query := `SELECT id, name, description FROM products`
	rows, err := r.pgClient.Query(query)
	if err != nil {
		return nil, err // Handle error appropriately
	}
	defer rows.Close()

	var products []product.Product
	for rows.Next() {
		var product product.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description); err != nil {
			return nil, err // Handle error appropriately
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err // Handle error appropriately
	}

	return products, nil
}

func (r *ProductRepository) SaveProduct(ctx context.Context, product *product.Product) (*product.Product, error) {
	query := `INSERT INTO products (id, name, description) VALUES ($1, $2, $3)`
	_, err := r.pgClient.Exec(query, product.ID, product.Name, product.Description)
	if err != nil {
		return nil, err
	}

	query = `SELECT id, name, description, created_at, updated_at FROM products WHERE id = $1`
	row := r.pgClient.QueryRow(query, product.ID)
	if err := row.Scan(&product.ID, &product.Name, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Product not found
		}
		return nil, err // Other error
	}

	// Return the saved product
	return product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, product *product.Product) (*product.Product, error) {
	query := `UPDATE products SET name = $1, description = $2 WHERE id = $3`
	_, err := r.pgClient.Exec(query, product.Name, product.Description, product.ID)
	if err != nil {
		return nil, err
	}

	query = `SELECT id, name, description, created_at, updated_at FROM products WHERE id = $1`
	row := r.pgClient.QueryRow(query, product.ID)
	if err := row.Scan(&product.ID, &product.Name, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Product not found
		}
		return nil, err // Other error
	}

	return product, nil
}

func (r *ProductRepository) DeleteProduct(ctx context.Context, productID string) error {
	query := `DELETE FROM products WHERE id = $1`
	_, err := r.pgClient.Exec(query, productID)
	if err != nil {
		return err // Handle error appropriately
	}
	return nil
}

func (r *ProductRepository) SearchProductsByName(ctx context.Context, name string) ([]product.Product, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM products WHERE name ILIKE $1`
	rows, err := r.pgClient.Query(query, "%"+name+"%")
	if err != nil {
		return nil, err // Handle error appropriately
	}
	defer rows.Close()

	var products []product.Product
	for rows.Next() {
		var product product.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Description, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, err // Handle error appropriately
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, err // Handle error appropriately
	}

	return products, nil
}
