package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
)

// ProductRepository handles product database operations
type ProductRepository struct {
	db *database.PostgresDB
}

// NewProductRepository creates a new ProductRepository
func NewProductRepository(db *database.PostgresDB) *ProductRepository {
	return &ProductRepository{db: db}
}

// Create creates a new product
func (r *ProductRepository) Create(ctx context.Context, input domain.ProductCreateInput) (*domain.Product, error) {
	// Set defaults
	isStockActive := true
	if input.IsStockActive != nil {
		isStockActive = *input.IsStockActive
	}
	currentStock := 0
	if input.CurrentStock != nil {
		currentStock = *input.CurrentStock
	}
	minStockAlert := 0
	if input.MinStockAlert != nil {
		minStockAlert = *input.MinStockAlert
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	// Handle empty strings for Unique constraints (convert to nil)
	if input.Barcode != nil && *input.Barcode == "" {
		input.Barcode = nil
	}
	if input.SKU != nil && *input.SKU == "" {
		input.SKU = nil
	}

	query := `
		INSERT INTO products (
			barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
	`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query,
		input.Barcode, input.SKU, input.Name, input.Description,
		input.CategoryID, input.LocationID, input.ConsignorID, input.Unit, input.BasePrice, input.CostPrice,
		isStockActive, currentStock, minStockAlert, input.MaxStock, input.ImageURL,
		isActive,
	).Scan(
		&product.ID, &product.Barcode, &product.SKU, &product.Name,
		&product.Description, &product.CategoryID, &product.LocationID, &product.ConsignorID, &product.Unit,
		&product.BasePrice, &product.CostPrice, &product.IsStockActive,
		&product.CurrentStock, &product.MinStockAlert, &product.MaxStock,
		&product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Create pricing tiers if provided
	if len(input.PricingTiers) > 0 {
		for _, tierInput := range input.PricingTiers {
			_, err := r.CreatePricingTier(ctx, product.ID, tierInput)
			if err != nil {
				return nil, fmt.Errorf("failed to create pricing tier: %w", err)
			}
		}
		// Reload pricing tiers
		product.PricingTiers, _ = r.GetPricingTiers(ctx, product.ID)
	}

	if product.LocationID != nil {
		locs, _ := r.GetLocationsBatch(ctx, []uuid.UUID{*product.LocationID})
		product.Location = locs[*product.LocationID]
	}

	return &product, nil
}

// GetByID retrieves a product by ID
func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		SELECT id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
		FROM products
		WHERE id = $1
	`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Barcode, &product.SKU, &product.Name,
		&product.Description, &product.CategoryID, &product.LocationID, &product.ConsignorID, &product.Unit,
		&product.BasePrice, &product.CostPrice, &product.IsStockActive,
		&product.CurrentStock, &product.MinStockAlert, &product.MaxStock,
		&product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Load pricing tiers
	product.PricingTiers, _ = r.GetPricingTiers(ctx, product.ID)

	if product.LocationID != nil {
		locs, _ := r.GetLocationsBatch(ctx, []uuid.UUID{*product.LocationID})
		product.Location = locs[*product.LocationID]
	}

	return &product, nil
}

// GetByBarcode retrieves a product by barcode
func (r *ProductRepository) GetByBarcode(ctx context.Context, barcode string) (*domain.Product, error) {
	query := `
		SELECT id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
		FROM products
		WHERE barcode = $1
	`

	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, barcode).Scan(
		&product.ID, &product.Barcode, &product.SKU, &product.Name,
		&product.Description, &product.CategoryID, &product.LocationID, &product.ConsignorID, &product.Unit,
		&product.BasePrice, &product.CostPrice, &product.IsStockActive,
		&product.CurrentStock, &product.MinStockAlert, &product.MaxStock,
		&product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get product by barcode: %w", err)
	}

	product.PricingTiers, _ = r.GetPricingTiers(ctx, product.ID)

	if product.LocationID != nil {
		locs, _ := r.GetLocationsBatch(ctx, []uuid.UUID{*product.LocationID})
		product.Location = locs[*product.LocationID]
	}

	return &product, nil
}

// List retrieves products with filtering and pagination
func (r *ProductRepository) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, int64, error) {
	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf(
			"(name ILIKE $%d OR barcode ILIKE $%d OR sku ILIKE $%d)",
			argIndex, argIndex, argIndex,
		))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.ConsignorID != nil {
		conditions = append(conditions, fmt.Sprintf("consignor_id = $%d", argIndex))
		args = append(args, *filter.ConsignorID)
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.IsStockActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_stock_active = $%d", argIndex))
		args = append(args, *filter.IsStockActive)
		argIndex++
	}

	if filter.LowStockOnly {
		conditions = append(conditions, "is_stock_active = true AND current_stock <= min_stock_alert")
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Validate sort options
	validSortFields := map[string]bool{
		"name": true, "created_at": true, "updated_at": true,
		"base_price": true, "current_stock": true,
	}
	sortBy := "name"
	if filter.SortBy != "" && validSortFields[filter.SortBy] {
		sortBy = filter.SortBy
	}
	sortOrder := "ASC"
	if strings.ToUpper(filter.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	// Main query
	query := fmt.Sprintf(`
		SELECT id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
		FROM products
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, perPage, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	products := make([]domain.Product, 0)
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.Barcode, &p.SKU, &p.Name, &p.Description, &p.CategoryID, &p.LocationID, &p.ConsignorID,
			&p.Unit, &p.BasePrice, &p.CostPrice, &p.IsStockActive, &p.CurrentStock,
			&p.MinStockAlert, &p.MaxStock, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	// Load pricing tiers and locations in batch (fixes N+1 query)
	if len(products) > 0 {
		productIDs := make([]uuid.UUID, len(products))
		var locIDs []uuid.UUID
		for i, p := range products {
			productIDs[i] = p.ID
			if p.LocationID != nil {
				locIDs = append(locIDs, *p.LocationID)
			}
		}

		tiersMap, _ := r.GetPricingTiersBatch(ctx, productIDs)
		locsMap, _ := r.GetLocationsBatch(ctx, locIDs)
		
		for i := range products {
			products[i].PricingTiers = tiersMap[products[i].ID]
			if products[i].LocationID != nil {
				products[i].Location = locsMap[*products[i].LocationID]
			}
		}
	}

	return products, total, rows.Err()
}

// Update updates a product
func (r *ProductRepository) Update(ctx context.Context, id uuid.UUID, input domain.ProductUpdateInput) (*domain.Product, error) {
	// Get current product
	product, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if input.Barcode != nil {
		setClauses = append(setClauses, fmt.Sprintf("barcode = $%d", argIndex))
		if *input.Barcode == "" {
			args = append(args, nil)
		} else {
			args = append(args, input.Barcode)
		}
		argIndex++
	}
	if input.SKU != nil {
		setClauses = append(setClauses, fmt.Sprintf("sku = $%d", argIndex))
		if *input.SKU == "" {
			args = append(args, nil)
		} else {
			args = append(args, input.SKU)
		}
		argIndex++
	}
	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *input.Name)
		argIndex++
	}
	if input.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, input.Description)
		argIndex++
	}
	if input.CategoryID != nil {
		setClauses = append(setClauses, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, input.CategoryID)
		argIndex++
	}
	if input.LocationID != nil {
		setClauses = append(setClauses, fmt.Sprintf("location_id = $%d", argIndex))
		args = append(args, input.LocationID)
		argIndex++
	}
	if input.ConsignorID != nil {
		setClauses = append(setClauses, fmt.Sprintf("consignor_id = $%d", argIndex))
		args = append(args, input.ConsignorID)
		argIndex++
	}
	if input.Unit != nil {
		setClauses = append(setClauses, fmt.Sprintf("unit = $%d", argIndex))
		args = append(args, *input.Unit)
		argIndex++
	}
	if input.BasePrice != nil {
		setClauses = append(setClauses, fmt.Sprintf("base_price = $%d", argIndex))
		args = append(args, *input.BasePrice)
		argIndex++
	}
	if input.CostPrice != nil {
		setClauses = append(setClauses, fmt.Sprintf("cost_price = $%d", argIndex))
		args = append(args, *input.CostPrice)
		argIndex++
	}
	if input.IsStockActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_stock_active = $%d", argIndex))
		args = append(args, *input.IsStockActive)
		argIndex++
	}
	if input.MinStockAlert != nil {
		setClauses = append(setClauses, fmt.Sprintf("min_stock_alert = $%d", argIndex))
		args = append(args, *input.MinStockAlert)
		argIndex++
	}
	if input.MaxStock != nil {
		setClauses = append(setClauses, fmt.Sprintf("max_stock = $%d", argIndex))
		args = append(args, input.MaxStock)
		argIndex++
	}
	if input.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *input.IsActive)
		argIndex++
	}
	if input.ImageURL != nil {
		setClauses = append(setClauses, fmt.Sprintf("image_url = $%d", argIndex))
		if *input.ImageURL == "" {
			args = append(args, nil)
		} else {
			args = append(args, input.ImageURL)
		}
		argIndex++
	}
	
	setClauses = append(setClauses, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	query := fmt.Sprintf(`
		UPDATE products
		SET %s
		WHERE id = $%d
		RETURNING id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex)

	args = append(args, id)

	var p domain.Product
	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&p.ID, &p.Barcode, &p.SKU, &p.Name, &p.Description, &p.CategoryID, &p.LocationID, &p.ConsignorID,
		&p.Unit, &p.BasePrice, &p.CostPrice, &p.IsStockActive, &p.CurrentStock,
		&p.MinStockAlert, &p.MaxStock, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	p.PricingTiers = product.PricingTiers // Keep same pricing tiers

	return &p, nil
}

// ToggleActive toggles the is_active status of a product
func (r *ProductRepository) ToggleActive(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	query := `
		UPDATE products
		SET is_active = NOT is_active, updated_at = NOW()
		WHERE id = $1
		RETURNING id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
	`

	var p domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Barcode, &p.SKU, &p.Name, &p.Description, &p.CategoryID, &p.LocationID, &p.ConsignorID,
		&p.Unit, &p.BasePrice, &p.CostPrice, &p.IsStockActive, &p.CurrentStock,
		&p.MinStockAlert, &p.MaxStock, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to toggle product active status: %w", err)
	}

	// Load pricing tiers
	p.PricingTiers, _ = r.GetPricingTiers(ctx, p.ID)

	return &p, nil
}




// UpdateStock updates the product stock
func (r *ProductRepository) UpdateStock(ctx context.Context, id uuid.UUID, newStock int) error {
	query := `UPDATE products SET current_stock = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, newStock, id)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	
	return nil
}

// UpdateImageURL updates the product image URL
func (r *ProductRepository) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL string) error {
	query := `UPDATE products SET image_url = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, imageURL, id)
	return err
}

// Delete soft deletes a product
func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1 AND is_active = true`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	
	return nil
}

// GetLowStockProducts gets products below minimum stock level
func (r *ProductRepository) GetLowStockProducts(ctx context.Context) ([]domain.LowStockProduct, error) {
	query := `
		SELECT id, barcode, sku, name, description, category_id, location_id, consignor_id, unit,
			base_price, cost_price, is_stock_active, current_stock,
			min_stock_alert, max_stock, image_url, is_active, created_at, updated_at
		FROM products
		WHERE is_stock_active = true 
			AND is_active = true 
			AND current_stock <= min_stock_alert
		ORDER BY (min_stock_alert - current_stock) DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products: %w", err)
	}
	defer rows.Close()

	var results []domain.LowStockProduct
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(
			&p.ID, &p.Barcode, &p.SKU, &p.Name, &p.Description, &p.CategoryID, &p.LocationID, &p.ConsignorID,
			&p.Unit, &p.BasePrice, &p.CostPrice, &p.IsStockActive, &p.CurrentStock,
			&p.MinStockAlert, &p.MaxStock, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		
		results = append(results, domain.LowStockProduct{
			Product:       p,
			DeficitAmount: p.MinStockAlert - p.CurrentStock,
		})
	}

	// Load locations in batch
	var locIDs []uuid.UUID
	for i := range results {
		if results[i].Product.LocationID != nil {
			locIDs = append(locIDs, *results[i].Product.LocationID)
		}
	}
	if len(locIDs) > 0 {
		locsMap, _ := r.GetLocationsBatch(ctx, locIDs)
		for i := range results {
			if results[i].Product.LocationID != nil {
				results[i].Product.Location = locsMap[*results[i].Product.LocationID]
			}
		}
	}

	return results, rows.Err()
}

// Pricing Tier methods

// CreatePricingTier creates a new pricing tier for a product
func (r *ProductRepository) CreatePricingTier(ctx context.Context, productID uuid.UUID, input domain.PricingTierInput) (*domain.PricingTier, error) {
	query := `
		INSERT INTO pricing_tiers (product_id, name, min_quantity, max_quantity, price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, product_id, name, min_quantity, max_quantity, price, is_active, created_at, updated_at
	`

	var tier domain.PricingTier
	err := r.db.QueryRowContext(ctx, query,
		productID, input.Name, input.MinQuantity, input.MaxQuantity, input.Price,
	).Scan(
		&tier.ID, &tier.ProductID, &tier.Name, &tier.MinQuantity,
		&tier.MaxQuantity, &tier.Price, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create pricing tier: %w", err)
	}

	return &tier, nil
}

// GetPricingTiers retrieves all pricing tiers for a product
func (r *ProductRepository) GetPricingTiers(ctx context.Context, productID uuid.UUID) ([]domain.PricingTier, error) {
	query := `
		SELECT id, product_id, name, min_quantity, max_quantity, price, is_active, created_at, updated_at
		FROM pricing_tiers
		WHERE product_id = $1 AND is_active = true
		ORDER BY min_quantity ASC
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tiers: %w", err)
	}
	defer rows.Close()

	var tiers []domain.PricingTier
	for rows.Next() {
		var t domain.PricingTier
		if err := rows.Scan(
			&t.ID, &t.ProductID, &t.Name, &t.MinQuantity,
			&t.MaxQuantity, &t.Price, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pricing tier: %w", err)
		}
		tiers = append(tiers, t)
	}

	return tiers, rows.Err()
}

// GetPricingTiersBatch retrieves pricing tiers for multiple products in a single query
// This fixes N+1 query issue when listing products
func (r *ProductRepository) GetPricingTiersBatch(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID][]domain.PricingTier, error) {
	if len(productIDs) == 0 {
		return make(map[uuid.UUID][]domain.PricingTier), nil
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, len(productIDs))
	for i, id := range productIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, product_id, name, min_quantity, max_quantity, price, is_active, created_at, updated_at
		FROM pricing_tiers
		WHERE product_id IN (%s) AND is_active = true
		ORDER BY product_id, min_quantity ASC
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tiers batch: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]domain.PricingTier)
	for rows.Next() {
		var t domain.PricingTier
		if err := rows.Scan(
			&t.ID, &t.ProductID, &t.Name, &t.MinQuantity,
			&t.MaxQuantity, &t.Price, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pricing tier: %w", err)
		}
		result[t.ProductID] = append(result[t.ProductID], t)
	}

	return result, rows.Err()
}

// GetPricingTier retrieves a single pricing tier by ID
func (r *ProductRepository) GetPricingTier(ctx context.Context, id uuid.UUID) (*domain.PricingTier, error) {
	query := `
		SELECT id, product_id, name, min_quantity, max_quantity, price, is_active, created_at, updated_at
		FROM pricing_tiers
		WHERE id = $1
	`
	var t domain.PricingTier
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.ProductID, &t.Name, &t.MinQuantity,
		&t.MaxQuantity, &t.Price, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get pricing tier: %w", err)
	}
	return &t, nil
}

// UpdatePricingTier updates a pricing tier
func (r *ProductRepository) UpdatePricingTier(ctx context.Context, tierID uuid.UUID, input domain.PricingTierInput) (*domain.PricingTier, error) {
	// Build dynamic query
	var setClauses []string
	var args []interface{}
	argIndex := 1

	if input.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, input.Name)
		argIndex++
	}
	if input.MinQuantity > 0 { // Assuming 0 implies no change? Or explicit 0? Input struct value type makes partial ambiguous. 
		// Actually PricingTierInput has MinQuantity int (value). So it's always 0 if not sent.
		// If user wants to update ONLY price, MinQuantity will be 0.
		// We should probably allow MinQuantity=0 if we assume full replacement OR partial?
		// Given we changed to partial, we should assume value types (int) are updated ONLY if non-zero?
		// But 0 might be valid (though unlikely for MinQuantity).
		// Let's assume > 0 for now. Or better: fetch existing first.
		setClauses = append(setClauses, fmt.Sprintf("min_quantity = $%d", argIndex))
		args = append(args, input.MinQuantity)
		argIndex++
	}
	if input.MaxQuantity != nil {
		setClauses = append(setClauses, fmt.Sprintf("max_quantity = $%d", argIndex))
		args = append(args, input.MaxQuantity)
		argIndex++
	}
	if input.Price > 0 {
		setClauses = append(setClauses, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, input.Price)
		argIndex++
	}

	if len(setClauses) == 0 {
		// Just return existing
		return r.GetPricingTier(ctx, tierID) // Need to implement GetPricingTier or just query
	}

	args = append(args, tierID)
	query := fmt.Sprintf(`
		UPDATE pricing_tiers 
		SET %s, updated_at = NOW()
		WHERE id = $%d
		RETURNING id, product_id, name, min_quantity, max_quantity, price, is_active, created_at, updated_at
	`, strings.Join(setClauses, ", "), argIndex)

	var tier domain.PricingTier
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&tier.ID, &tier.ProductID, &tier.Name, &tier.MinQuantity,
		&tier.MaxQuantity, &tier.Price, &tier.IsActive, &tier.CreatedAt, &tier.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update pricing tier: %w", err)
	}

	return &tier, nil
}

// PublicProductFilter is a simplified filter for public product listing
type PublicProductFilter struct {
	Search     *string
	CategoryID *uuid.UUID
	IsActive   *bool
	Page       int
	PerPage    int
	SortBy     string
	SortOrder  string
}

// ListPublic retrieves active products with category info for the public landing page
func (r *ProductRepository) ListPublic(ctx context.Context, filter PublicProductFilter) ([]domain.Product, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("p.name ILIKE $%d", argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if filter.CategoryID != nil {
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("p.is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s", whereClause)
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Pagination
	page := filter.Page
	if page < 1 {
		page = 1
	}
	perPage := filter.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	validSortFields := map[string]bool{"name": true, "base_price": true, "created_at": true}
	sortBy := "p.name"
	if filter.SortBy != "" && validSortFields[filter.SortBy] {
		sortBy = "p." + filter.SortBy
	}
	sortOrder := "ASC"
	if strings.ToUpper(filter.SortOrder) == "DESC" {
		sortOrder = "DESC"
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.unit, p.base_price, p.image_url,
			p.category_id, c.name, c.description
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id AND c.is_active = true
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortOrder, argIndex, argIndex+1)

	args = append(args, perPage, (page-1)*perPage)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list public products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var p domain.Product
		var catName *string
		var catDesc *string
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Unit, &p.BasePrice, &p.ImageURL,
			&p.CategoryID, &catName, &catDesc,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}
		if p.CategoryID != nil && catName != nil {
			p.Category = &domain.Category{
				ID:          *p.CategoryID,
				Name:        *catName,
				Description: catDesc,
			}
		}
		products = append(products, p)
	}

	// Load pricing tiers in batch
	if len(products) > 0 {
		productIDs := make([]uuid.UUID, len(products))
		for i, p := range products {
			productIDs[i] = p.ID
		}
		tiersMap, err := r.GetPricingTiersBatch(ctx, productIDs)
		if err == nil {
			for i := range products {
				products[i].PricingTiers = tiersMap[products[i].ID]
			}
		}
	}

	return products, total, rows.Err()
}

// DeletePricingTier soft deletes a pricing tier
func (r *ProductRepository) DeletePricingTier(ctx context.Context, tierID uuid.UUID) error {
	query := `UPDATE pricing_tiers SET is_active = false, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, tierID)
	if err != nil {
		return fmt.Errorf("failed to delete pricing tier: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	
	return nil
}

// GetLocationsBatch retrieves locations for multiple IDs in a single query
func (r *ProductRepository) GetLocationsBatch(ctx context.Context, locationIDs []uuid.UUID) (map[uuid.UUID]*domain.Location, error) {
	if len(locationIDs) == 0 {
		return make(map[uuid.UUID]*domain.Location), nil
	}

	// Deduplicate IDs
	uniqueIDs := make(map[uuid.UUID]bool)
	var filteredIDs []uuid.UUID
	for _, id := range locationIDs {
		if !uniqueIDs[id] {
			uniqueIDs[id] = true
			filteredIDs = append(filteredIDs, id)
		}
	}

	placeholders := make([]string, len(filteredIDs))
	args := make([]interface{}, len(filteredIDs))
	for i, id := range filteredIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, name, category, x_coordinate, y_coordinate, z_coordinate,
			width, depth, height, description, is_active, created_at, updated_at
		FROM locations
		WHERE id IN (%s) AND is_active = true
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get locations batch: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]*domain.Location)
	for rows.Next() {
		var loc domain.Location
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Category,
			&loc.XCoordinate, &loc.YCoordinate, &loc.ZCoordinate,
			&loc.Width, &loc.Depth, &loc.Height,
			&loc.Description, &loc.IsActive, &loc.CreatedAt, &loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan location: %w", err)
		}
		locCopy := loc // Avoid pointer to loop variable issue
		result[loc.ID] = &locCopy
	}

	return result, rows.Err()
}
