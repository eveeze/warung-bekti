package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/repository"
)

func setupTestDB(t *testing.T) (*database.PostgresDB, context.Context) {
	dbCfg := &config.DatabaseConfig{
		Host:            "127.0.0.1",
		Port:            "5435",
		User:            "warung",
		Password:        "warung_secret",
		DBName:          "warung_db",
		SSLMode:         "disable",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	db, err := database.NewPostgres(dbCfg)
	if err != nil {
		t.Fatalf("failed to connect live test database: %v", err)
	}
	return db, context.Background()
}

func TestAllRepositoriesLive(t *testing.T) {
	db, ctx := setupTestDB(t)
	defer db.Close()

	// 1. Category Repository
	catRepo := repository.NewCategoryRepository(db)
	cat := &domain.Category{
		Name: "Repo Cat " + uuid.New().String()[:8],
	}
	err := catRepo.Create(ctx, cat)
	if err != nil {
		t.Fatalf("catRepo.Create failed: %v", err)
	}
	_, _ = catRepo.FindByID(ctx, cat.ID)
	_, _ = catRepo.FindAll(ctx, nil)
	cat.Name = "Updated Cat Name"
	_ = catRepo.Update(ctx, cat)

	// 2. Location Repository
	locRepo := repository.NewLocationRepository(db)
	loc, err := locRepo.Create(ctx, domain.LocationCreateInput{
		Name:        "Repo Shelf " + uuid.New().String()[:4],
		Category:    "shelf",
		Width:       2.0,
		Depth:       1.0,
		Height:      2.0,
		XCoordinate: 0,
		YCoordinate: 0,
		ZCoordinate: 0,
	})
	if err == nil && loc != nil {
		_, _ = locRepo.GetByID(ctx, loc.ID)
		_, _ = locRepo.ListAll(ctx)
		updatedLocName := "Updated Shelf"
		updatedCat := domain.LocationTypeShelf
		desc := "Loc desc"
		w, d, h := 3.0, 2.0, 4.0
		x, y, z := 1.0, 2.0, 3.0
		act := true
		_, _ = locRepo.Update(ctx, loc.ID, domain.LocationUpdateInput{
			Name:        &updatedLocName,
			Category:    &updatedCat,
			Description: &desc,
			Width:       &w,
			Depth:       &d,
			Height:      &h,
			XCoordinate: &x,
			YCoordinate: &y,
			ZCoordinate: &z,
			IsActive:    &act,
		})
		_ = locRepo.Delete(ctx, loc.ID)
	}

	// 3. User Repository
	userRepo := repository.NewUserRepository(db)
	user := &domain.User{
		Name:         "Repo User",
		Email:        "repo_" + uuid.New().String()[:8] + "@warung.com",
		PasswordHash: "hash123",
		Role:         domain.RoleCashier,
		IsActive:     true,
	}
	if err := userRepo.Create(ctx, user); err == nil {
		_, _ = userRepo.GetByID(ctx, user.ID)
		_, _ = userRepo.GetByEmail(ctx, user.Email)
		_, _, _ = userRepo.List(ctx, domain.UserListParams{Page: 1, PerPage: 10})
		user.Name = "Repo User Updated"
		_ = userRepo.Update(ctx, user)
		_ = userRepo.UpdateLastLogin(ctx, user.ID)
		_ = userRepo.Delete(ctx, user.ID)
	}

	// 4. Customer Repository
	custRepo := repository.NewCustomerRepository(db)
	var creditLimit int64 = 500000
	phone := "081" + uuid.New().String()[:9]
	cust, err := custRepo.Create(ctx, domain.CustomerCreateInput{
		Name:        "Repo Cust " + uuid.New().String()[:8],
		Phone:       &phone,
		CreditLimit: &creditLimit,
	})
	if err == nil && cust != nil {
		_, _ = custRepo.GetByID(ctx, cust.ID)
		_, _, _ = custRepo.List(ctx, domain.CustomerFilter{Page: 1, PerPage: 10})
		_, _ = custRepo.GetCustomersWithDebt(ctx)
		_, _ = custRepo.GetTotalOutstandingDebt(ctx)
		newLimit := int64(750000)
		newName := "Cust Updated"
		newAddress := "Jl. Baru No. 1"
		newNotes := "Notes baru"
		isActiveCust := true
		_, _ = custRepo.Update(ctx, cust.ID, domain.CustomerUpdateInput{
			Name:        &newName,
			Phone:       &phone,
			Address:     &newAddress,
			Notes:       &newNotes,
			CreditLimit: &newLimit,
			IsActive:    &isActiveCust,
		})
		_ = custRepo.AddDebt(ctx, cust.ID, 10000)
		_ = custRepo.SubtractDebt(ctx, cust.ID, 5000)
		_ = custRepo.UpdateDebt(ctx, cust.ID, 15000)
		_ = custRepo.Delete(ctx, cust.ID)
	}

	// 5. Product Repository
	prodRepo := repository.NewProductRepository(db)
	barcode := "BAR" + uuid.New().String()[:10]
	isStock := true
	currentStock := 50
	minAlert := 5
	prod, err := prodRepo.Create(ctx, domain.ProductCreateInput{
		Name:          "Repo Prod " + uuid.New().String()[:8],
		Barcode:       &barcode,
		Unit:          "pcs",
		BasePrice:     15000,
		CostPrice:     10000,
		IsStockActive: &isStock,
		CurrentStock:  &currentStock,
		MinStockAlert: &minAlert,
	})
	if err == nil && prod != nil {
		_, _ = prodRepo.GetByID(ctx, prod.ID)
		_, _ = prodRepo.GetByBarcode(ctx, barcode)
		searchStr := "Repo"
		isActiveVal := true
		isStockVal := true
		_, _, _ = prodRepo.List(ctx, domain.ProductFilter{
			Search:        &searchStr,
			CategoryID:    prod.CategoryID,
			IsActive:      &isActiveVal,
			IsStockActive: &isStockVal,
			LowStockOnly:  true,
			Page:          1,
			PerPage:       10,
		})
		_, _ = prodRepo.GetLowStockProducts(ctx)
		_, _ = prodRepo.ToggleActive(ctx, prod.ID)
		_, _ = prodRepo.ToggleActive(ctx, prod.ID)

		// Full update on product
		newProdName := "Repo Prod Updated Full"
		newDesc := "Updated Description"
		newUnit := "box"
		newBasePrice := int64(18000)
		newCostPrice := int64(12000)
		newMinAlert := 8
		newMaxStock := 200
		newImgURL := "https://cdn.warung.com/new.jpg"
		emptyImgURL := ""
		_, _ = prodRepo.Update(ctx, prod.ID, domain.ProductUpdateInput{
			Name:          &newProdName,
			Description:   &newDesc,
			Unit:          &newUnit,
			BasePrice:     &newBasePrice,
			CostPrice:     &newCostPrice,
			IsStockActive: &isStockVal,
			MinStockAlert: &newMinAlert,
			MaxStock:      &newMaxStock,
			IsActive:      &isActiveVal,
			ImageURL:      &newImgURL,
		})
		_, _ = prodRepo.Update(ctx, prod.ID, domain.ProductUpdateInput{
			ImageURL: &emptyImgURL,
		})
		_ = prodRepo.UpdateStock(ctx, prod.ID, 88)

		// Pricing tier
		maxQty := 30
		tierName := "Grosir"
		tier, _ := prodRepo.CreatePricingTier(ctx, prod.ID, domain.PricingTierInput{
			Name:        &tierName,
			MinQuantity: 10,
			MaxQuantity: &maxQty,
			Price:       14000,
		})
		if tier != nil {
			tier.Price = 13500
			_, _ = prodRepo.GetPricingTier(ctx, tier.ID)
			_, _ = prodRepo.UpdatePricingTier(ctx, tier.ID, domain.PricingTierInput{
				Price:       13500,
				MinQuantity: 10,
			})
			_ = prodRepo.DeletePricingTier(ctx, tier.ID)
		}
		_ = prodRepo.UpdateImageURL(ctx, prod.ID, "https://cdn.warung.com/img.jpg")
		_, _ = prodRepo.GetLocationsBatch(ctx, nil)
		if loc != nil {
			_, _ = prodRepo.GetLocationsBatch(ctx, []uuid.UUID{loc.ID})
			_, _ = prodRepo.Update(ctx, prod.ID, domain.ProductUpdateInput{
				LocationID: &loc.ID,
			})
		}
		if cat != nil {
			_, _ = prodRepo.Update(ctx, prod.ID, domain.ProductUpdateInput{
				CategoryID: &cat.ID,
			})
		}
		newSKU := "SKU-999"
		_, _ = prodRepo.Update(ctx, prod.ID, domain.ProductUpdateInput{
			SKU: &newSKU,
		})
		_ = prodRepo.Delete(ctx, prod.ID)
	}

	// 6. Cash Flow Repository
	cfRepo := repository.NewCashFlowRepository(db)
	_, _ = cfRepo.GetCategories(ctx)
	notes := "Buka Laci Repo Test"
	session, err := cfRepo.OpenDrawer(ctx, domain.OpenDrawerInput{
		OpeningBalance: 100000,
		OpenedBy:       "RepoTester",
		Notes:          &notes,
	})
	if err != nil {
		session, _ = cfRepo.GetCurrentSession(ctx)
	}
	if session != nil {
		desc := "Uang Masuk Tambahan"
		_, _ = cfRepo.RecordCashFlow(ctx, nil, domain.CashFlowInput{
			Type:        domain.CashFlowTypeIncome,
			Amount:      25000,
			Description: &desc,
			CreatedBy:   "RepoTester",
		}, &session.ID, nil, nil)
		_, _, _ = cfRepo.ListCashFlows(ctx, domain.CashFlowFilter{Page: 1, PerPage: 10})
	}

	// 7. Inventory Repository
	invRepo := repository.NewInventoryRepository(db)
	if prod != nil {
		notes := "Restock Masuk"
		_, _ = invRepo.Restock(ctx, domain.RestockInput{
			ProductID:   prod.ID,
			Quantity:    20,
			CostPerUnit: 10000,
			Notes:       &notes,
		})
		_, _ = invRepo.Adjust(ctx, domain.StockAdjustmentInput{
			ProductID: prod.ID,
			Quantity:  -2,
			Reason:    "Pecah",
		})
		_, _, _ = invRepo.GetByProduct(ctx, prod.ID, domain.StockMovementFilter{Page: 1, PerPage: 10})
	}
	_, _ = invRepo.GetStockReport(ctx)

	// 8. Transaction Repository
	txRepo := repository.NewTransactionRepository(db)
	var createdTx *domain.Transaction
	if prod != nil {
		cashier := "RepoCashier"
		_ = db.WithTransaction(ctx, func(tx *sql.Tx) error {
			tRecord := &domain.Transaction{
				CustomerID:    nil,
				PaymentMethod: domain.PaymentMethodCash,
				AmountPaid:    50000,
				Status:        domain.TransactionStatusCompleted,
				CashierName:   &cashier,
				Items: []domain.TransactionItem{
					{
						ProductID:   prod.ID,
						Quantity:    1,
						UnitPrice:   prod.BasePrice,
						CostPrice:   prod.CostPrice,
						Subtotal:    prod.BasePrice,
						TotalAmount: prod.BasePrice,
					},
				},
				Subtotal:    prod.BasePrice,
				TotalAmount: prod.BasePrice,
			}
			err := txRepo.Create(ctx, tx, tRecord)
			if err == nil {
				createdTx = tRecord
			}
			return err
		})
	}
	if createdTx != nil {
		_, _ = txRepo.GetByID(ctx, createdTx.ID)
		_, _, _ = txRepo.List(ctx, domain.TransactionFilter{Page: 1, PerPage: 10})
		today := time.Now().Format("2006-01-02")
		_, _, _ = txRepo.GetDailySales(ctx, today)
		_, _ = txRepo.GetDailyProfit(ctx, today)
		_, _ = txRepo.GetHourlySales(ctx, today)
		_, _ = txRepo.GetTopProducts(ctx, today, 5)
		_ = txRepo.UpdateStatus(ctx, createdTx.ID, domain.TransactionStatusCancelled)
	}

	// 9. Kasbon Repository
	kasbonRepo := repository.NewKasbonRepository(db)
	if cust != nil {
		paymentNotes := "Bayar Cicilan"
		_, _ = kasbonRepo.CreatePayment(ctx, domain.KasbonPaymentInput{
			CustomerID: cust.ID,
			Amount:     10000,
			Notes:      &paymentNotes,
		})
		_, _, _ = kasbonRepo.GetByCustomer(ctx, cust.ID, domain.KasbonFilter{Page: 1, PerPage: 10})
		_, _ = kasbonRepo.GetSummary(ctx, cust.ID)
		_, _ = kasbonRepo.GetReport(ctx)
	}

	// 10. POS Repository (Held Carts & Refunds)
	posRepo := repository.NewPOSRepository(db)
	if prod != nil {
		notes := "Cart Hold"
		tester := "Tester"
		cart := &domain.HeldCart{
			Items:  []domain.HeldCartItem{{ProductID: prod.ID, Quantity: 2}},
			Notes:  &notes,
			HeldBy: &tester,
		}
		err := posRepo.HoldCart(ctx, cart)
		if err == nil {
			_, _ = posRepo.GetHeldCart(ctx, cart.ID)
			_, _ = posRepo.ListHeldCarts(ctx)
			_ = posRepo.UpdateCartStatus(ctx, cart.ID, domain.HeldCartStatusDiscarded, &tester)
		}

		if createdTx != nil && len(createdTx.Items) > 0 {
			refundNum := "REF-" + uuid.New().String()[:8]
			reqBy := "Tester"
			refRecord := &domain.RefundRecord{
				RefundNumber:      refundNum,
				TransactionID:     createdTx.ID,
				TotalRefundAmount: 10000,
				RefundMethod:      "cash",
				Status:            domain.RefundStatusCompleted,
				Reason:            "Barang rusak",
				RequestedBy:       &reqBy,
				Items: []domain.RefundItem{
					{
						TransactionItemID: createdTx.Items[0].ID,
						ProductID:         prod.ID,
						ProductName:       prod.Name,
						Quantity:          1,
						UnitPrice:         prod.BasePrice,
						RefundAmount:      prod.BasePrice,
						Restock:           true,
					},
				},
			}
			if err := posRepo.CreateRefund(ctx, refRecord); err == nil {
				_, _ = posRepo.GetRefund(ctx, refRecord.ID)
			}
		}
	}

	// 11. Refillable Repository
	refillRepo := repository.NewRefillableRepository(db)
	if prod != nil {
		rc := &domain.RefillableContainer{
			ProductID:     prod.ID,
			ContainerType: "galon",
			EmptyCount:    3,
			FullCount:     7,
		}
		err := refillRepo.Create(ctx, rc)
		if err == nil {
			_, _ = refillRepo.GetContainers(ctx)
			_, _ = refillRepo.GetByProductID(ctx, prod.ID)
			_ = db.WithTransaction(ctx, func(tx *sql.Tx) error {
				_ = refillRepo.UpdateContainerStock(ctx, tx, rc.ID, 1, -1)
				return refillRepo.RecordMovement(ctx, tx, &domain.ContainerMovement{
					ContainerID: rc.ID,
					Type:        domain.ContainerMovementAdjustment,
					EmptyChange: 1,
					FullChange:  -1,
				})
			})
			_, _, _ = refillRepo.GetMovements(ctx, rc.ID, 10, 0)
		}
	}

	// 12. Consignment Repository
	consignRepo := repository.NewConsignmentRepository(db)
	con, _ := consignRepo.CreateConsignor(ctx, domain.CreateConsignorInput{
		Name: "Penitip Repo " + uuid.New().String()[:4],
	})
	if con != nil {
		_, _ = consignRepo.ListConsignors(ctx)
		updatedName := "Penitip Updated"
		_, _ = consignRepo.UpdateConsignor(ctx, con.ID, domain.UpdateConsignorInput{Name: &updatedName})

		if prod != nil {
			settlement := &domain.ConsignmentSettlement{
				ConsignorID:      con.ID,
				PeriodStart:      time.Now().Add(-24 * time.Hour),
				PeriodEnd:        time.Now(),
				TotalSales:       100000,
				CommissionAmount: 15000,
				ConsignorAmount:  85000,
				Status:           domain.SettlementStatusConfirmed,
				Items: []domain.ConsignmentSettlementItem{
					{
						ProductID:        prod.ID,
						ProductName:      prod.Name,
						QuantitySold:     5,
						UnitPrice:        20000,
						TotalSales:       100000,
						CommissionRate:   15.0,
						CommissionAmount: 15000,
						ConsignorAmount:  85000,
					},
				},
			}
			_ = consignRepo.CreateSettlement(ctx, settlement)
		}
		_ = consignRepo.DeleteConsignor(ctx, con.ID)
	}

	// 13. Notification Repository
	notifRepo := repository.NewNotificationRepository(db)
	notifUID := uuid.New()
	notif := &repository.Notification{
		UserID:  &notifUID,
		Title:   "Notif Title",
		Message: "Notif Message",
		Type:    "low_stock",
	}
	err = notifRepo.Create(ctx, notif)
	if err == nil {
		_, _ = notifRepo.GetByUserID(ctx, notifUID, 10, 0)
		_ = notifRepo.MarkAsRead(ctx, notif.ID)
		_ = notifRepo.MarkAllAsRead(ctx, notifUID)
	}

	// 14. Payment Repository
	payRepo := repository.NewPaymentRepository(db)
	var payRecord *domain.PaymentRecord
	if createdTx != nil {
		payRecord = &domain.PaymentRecord{
			TransactionID: createdTx.ID,
			OrderID:       "ORDER-" + uuid.New().String()[:8],
			GrossAmount:   createdTx.TotalAmount,
			Currency:      "IDR",
			Status:        domain.PaymentStatusPending,
		}
		_ = payRepo.Create(ctx, nil, payRecord)
		_, _ = payRepo.GetByID(ctx, payRecord.ID)
		_, _ = payRepo.GetByTransactionID(ctx, createdTx.ID)
		_, _ = payRepo.GetByOrderID(ctx, payRecord.OrderID)
		_ = payRepo.UpdateStatus(ctx, payRecord.ID, domain.PaymentStatusSettlement, nil)
		_ = payRepo.UpdateSnapToken(ctx, payRecord.ID, "snap-token-123", "https://app.sandbox.midtrans.com/snap/v2/vtweb/123")
	}

	// 15. Audit Repository
	auditRepo := repository.NewAuditRepository(db)
	auditReqID := "req-123"
	auditNotes := "Audit test note"
	_ = auditRepo.Log(ctx, &repository.AuditLog{
		Action:     "TEST_ACTION",
		EntityType: "test",
		RequestID:  &auditReqID,
		Notes:      &auditNotes,
	})

	// Clean up / close drawer
	if session != nil {
		closeNotes := "Selesai test"
		_, _ = cfRepo.CloseDrawer(ctx, domain.CloseDrawerInput{
			SessionID:      session.ID,
			ClosingBalance: 125000,
			ClosedBy:       "RepoTester",
			Notes:          &closeNotes,
		})
	}
}
