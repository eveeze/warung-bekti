package service

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/domain"
	"github.com/eveeze/warung-backend/internal/pkg/password"
	"github.com/eveeze/warung-backend/internal/platform/queue"
	"github.com/eveeze/warung-backend/internal/repository"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// MockUserRepository implements domain.UserRepository
type MockUserRepository struct {
	mu    sync.Mutex
	users map[uuid.UUID]*domain.User
	byEmail map[string]*domain.User
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:   make(map[uuid.UUID]*domain.User),
		byEmail: make(map[string]*domain.User),
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.byEmail[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[user.ID] = user
	m.byEmail[user.Email] = user
	return nil
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

func (m *MockUserRepository) List(ctx context.Context, params domain.UserListParams) ([]domain.User, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		list = append(list, *u)
	}
	return list, int64(len(list)), nil
}

func (m *MockUserRepository) Count(ctx context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.users)), nil
}

func TestEventService(t *testing.T) {
	svc := NewEventService()
	ch := svc.Subscribe()

	go func() {
		svc.Publish(EventStockUpdate, map[string]interface{}{"product_id": "123", "stock": 10})
	}()

	select {
	case ev := <-ch:
		if ev.Type != EventStockUpdate {
			t.Errorf("expected stock_update event, got %s", ev.Type)
		}
	case <-time.After(time.Second):
		t.Errorf("timeout waiting for event")
	}

	svc.Unsubscribe(ch)
}

func TestCacheService(t *testing.T) {
	// Test nil redis client error handling
	nilSvc := NewCacheService(nil)
	ctx := context.Background()
	if err := nilSvc.Get(ctx, "k", nil); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("expected ErrCacheUnavailable")
	}
	if err := nilSvc.Set(ctx, "k", "v", time.Minute); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("expected ErrCacheUnavailable")
	}
	if err := nilSvc.Delete(ctx, "k"); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("expected ErrCacheUnavailable")
	}
	if err := nilSvc.InvalidatePattern(ctx, "k*"); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("expected ErrCacheUnavailable")
	}
	if _, err := nilSvc.Exists(ctx, "k"); !errors.Is(err, ErrCacheUnavailable) {
		t.Errorf("expected ErrCacheUnavailable")
	}

	// Test with local live Redis on 6381
	redisCfg := &config.RedisConfig{Host: "127.0.0.1", Port: "6381"}
	client, err := database.NewRedis(redisCfg)
	if err != nil {
		t.Logf("skipping live redis cache test: %v", err)
		return
	}
	defer client.Close()

	svc := NewCacheService(client)
	testKey := "test:cache:key"
	testVal := map[string]string{"foo": "bar"}

	if err := svc.Set(ctx, testKey, testVal, time.Minute); err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	var retrieved map[string]string
	if err := svc.Get(ctx, testKey, &retrieved); err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}
	if retrieved["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", retrieved)
	}

	exists, err := svc.Exists(ctx, testKey)
	if err != nil || !exists {
		t.Errorf("expected exists=true, got %v, err: %v", exists, err)
	}

	_ = svc.Set(ctx, "test:cache:p1", "v1", time.Minute)
	_ = svc.Set(ctx, "test:cache:p2", "v2", time.Minute)
	if err := svc.InvalidatePattern(ctx, "test:cache:*"); err != nil {
		t.Errorf("failed to invalidate pattern: %v", err)
	}

	if err := svc.Delete(ctx, testKey); err != nil {
		t.Errorf("failed to delete cache: %v", err)
	}
}

func TestAuthService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-auth-secret-12345",
			ExpirationHours: 24,
			Issuer:          "warung-test",
		},
	}
	authSvc := NewAuthService(mockRepo, cfg)
	ctx := context.Background()

	// 1. Register admin should be rejected
	_, err := authSvc.Register(ctx, domain.RegisterRequest{
		Name:     "Admin Fake",
		Email:    "admin_fake@warung.com",
		Password: "password123",
		Role:     domain.RoleAdmin,
	})
	if err == nil {
		t.Errorf("expected error registering admin role")
	}

	// 2. Register cashier should succeed
	regResp, err := authSvc.Register(ctx, domain.RegisterRequest{
		Name:     "Kasir Budi",
		Email:    "budi@warung.com",
		Password: "password123",
		Role:     domain.RoleCashier,
	})
	if err != nil {
		t.Fatalf("expected register cashier success, got: %v", err)
	}
	if regResp.Email != "budi@warung.com" {
		t.Errorf("unexpected registered email")
	}
	_, _, _ = authSvc.generateTokens(regResp)

	// 3. Register duplicate email should be rejected
	_, err = authSvc.Register(ctx, domain.RegisterRequest{
		Name:     "Budi Duplicate",
		Email:    "budi@warung.com",
		Password: "password123",
		Role:     domain.RoleCashier,
	})
	if err == nil {
		t.Errorf("expected error registering duplicate email")
	}

	// 4. Login wrong password should fail
	_, err = authSvc.Login(ctx, domain.LoginRequest{
		Email:    "budi@warung.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Errorf("expected error for wrong password")
	}

	// 5. Login non-existent email should fail
	_, err = authSvc.Login(ctx, domain.LoginRequest{
		Email:    "nonexistent@warung.com",
		Password: "password123",
	})
	if err == nil {
		t.Errorf("expected error for non-existent user")
	}

	// 6. Login valid should succeed
	loginResp, err := authSvc.Login(ctx, domain.LoginRequest{
		Email:    "budi@warung.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("expected login success, got: %v", err)
	}
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Errorf("expected access and refresh tokens")
	}

	// 7. Refresh token valid should succeed
	refreshResp, err := authSvc.RefreshToken(ctx, loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("expected refresh token success, got: %v", err)
	}
	if refreshResp.AccessToken == "" {
		t.Errorf("expected access token from refresh")
	}

	// 8. Refresh invalid token should fail
	_, err = authSvc.RefreshToken(ctx, "invalid.refresh.token")
	if err == nil {
		t.Errorf("expected error for invalid refresh token")
	}

	// 9. Login inactive user should fail
	inactiveHash, _ := password.Hash("password123")
	inactiveUser := &domain.User{
		Name:         "Inactive User",
		Email:        "inactive@warung.com",
		PasswordHash: inactiveHash,
		Role:         domain.RoleCashier,
		IsActive:     false,
	}
	_ = mockRepo.Create(ctx, inactiveUser)

	_, err = authSvc.Login(ctx, domain.LoginRequest{
		Email:    "inactive@warung.com",
		Password: "password123",
	})
	if err == nil {
		t.Errorf("expected error logging in inactive user")
	}
}

func TestUserService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	userSvc := NewUserService(mockRepo)
	ctx := context.Background()

	// Create
	created, err := userSvc.CreateUser(ctx, domain.RegisterRequest{
		Name:     "Staff A",
		Email:    "staffA@warung.com",
		Password: "password123",
		Role:     domain.RoleInventory,
	})
	if err != nil {
		t.Fatalf("create user failed: %v", err)
	}

	// GetByID
	got, err := userSvc.GetByID(ctx, created.ID)
	if err != nil || got.Email != "staffA@warung.com" {
		t.Errorf("get user failed: %v", err)
	}

	// List
	users, total, err := userSvc.ListUsers(ctx, domain.UserListParams{Page: 1, PerPage: 10})
	if err != nil || total != 1 || len(users) != 1 {
		t.Errorf("list users failed: %v", err)
	}

	// Update
	nameUpdated := "Staff A Updated"
	active := true
	updated, err := userSvc.UpdateUser(ctx, created.ID, domain.UpdateUserRequest{
		Name:     &nameUpdated,
		IsActive: &active,
	})
	if err != nil || updated.Name != nameUpdated {
		t.Errorf("update user failed: %v", err)
	}

	// Delete
	if err := userSvc.DeleteUser(ctx, created.ID); err != nil {
		t.Errorf("delete user failed: %v", err)
	}

	// Non-existent user errors
	nonUID := uuid.New()
	_, _ = userSvc.GetByID(ctx, nonUID)
	_, _ = userSvc.UpdateUser(ctx, nonUID, domain.UpdateUserRequest{})
	_ = userSvc.DeleteUser(ctx, nonUID)
}

func TestCategoryService(t *testing.T) {
	// Use live local DB on 5435
	dbCfg := &config.DatabaseConfig{
		Host:     "127.0.0.1",
		Port:     "5435",
		User:     "warung",
		Password: "warung_secret",
		DBName:   "warung_db",
		SSLMode:  "disable",
	}
	db, err := database.NewPostgres(dbCfg)
	if err != nil {
		t.Logf("skipping live category service test: %v", err)
		return
	}
	defer db.Close()

	catRepo := repository.NewCategoryRepository(db)
	catSvc := NewCategoryService(catRepo)
	ctx := context.Background()

	created, err := catSvc.Create(ctx, domain.CategoryCreateInput{
		Name: "Unit Test Cat",
	})
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}

	got, err := catSvc.Get(ctx, created.ID)
	if err != nil || got == nil {
		t.Errorf("failed to get category: %v", err)
	}

	list, err := catSvc.List(ctx, nil)
	if err != nil || len(list) == 0 {
		t.Errorf("failed to list categories: %v", err)
	}

	desc := "Updated Description"
	updated, err := catSvc.Update(ctx, created.ID, domain.CategoryUpdateInput{
		Description: &desc,
	})
	if err != nil || updated == nil {
		t.Errorf("failed to update category: %v", err)
	}

	if err := catSvc.Delete(ctx, created.ID); err != nil {
		t.Errorf("failed to delete category: %v", err)
	}
}

func TestAllOtherServicesLive(t *testing.T) {
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
		t.Fatalf("failed to connect live database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// Repositories
	productRepo := repository.NewProductRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)
	kasbonRepo := repository.NewKasbonRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db)
	cashFlowRepo := repository.NewCashFlowRepository(db)
	posRepo := repository.NewPOSRepository(db)
	stockOpnameRepo := repository.NewStockOpnameRepository(db)
	refillableRepo := repository.NewRefillableRepository(db)
	consignmentRepo := repository.NewConsignmentRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	// 1. CashFlowService
	cfSvc := NewCashFlowService(db, cashFlowRepo)
	_, _ = cfSvc.GetCategories(ctx)

	// Ensure open drawer session
	openNotes := "Unit Test Modal Awal"
	session, err := cfSvc.OpenDrawer(ctx, domain.OpenDrawerInput{
		OpeningBalance: 150000,
		OpenedBy:       "Tester",
		Notes:          &openNotes,
	})
	if err != nil {
		// Maybe already open
		session, _ = cfSvc.GetCurrentSession(ctx)
	}
	if session != nil {
		desc := "Beli alat tulis"
		_, _ = cfSvc.RecordCashFlow(ctx, domain.CashFlowInput{
			Type:        domain.CashFlowTypeExpense,
			Amount:      10000,
			Description: &desc,
			CreatedBy:   "Tester",
		})
		_, _, _ = cfSvc.ListCashFlows(ctx, domain.CashFlowFilter{Page: 1, PerPage: 10})
	}

	// 2. RefillableService
	refillSvc := NewRefillableService(db, refillableRepo)
	// Create test product for container
	pName := "Air Mineral Galon Test"
	isStock := true
	stockVal := 20
	tp, _ := productRepo.Create(ctx, domain.ProductCreateInput{
		Name:          pName,
		Unit:          "galon",
		BasePrice:     20000,
		CostPrice:     15000,
		IsStockActive: &isStock,
		CurrentStock:  &stockVal,
	})
	if tp != nil {
		rc, _ := refillSvc.CreateContainer(ctx, domain.RefillableContainer{
			ContainerType: "galon",
			ProductID:     tp.ID,
			EmptyCount:    5,
			FullCount:     15,
		})
		if rc != nil {
			_, _ = refillSvc.GetContainers(ctx)
			createdByName := "Tester"
			_, _ = refillSvc.AdjustStock(ctx, domain.ContainerMovement{
				ContainerID: rc.ID,
				EmptyChange: 2,
				FullChange:  -2,
				CreatedBy:   &createdByName,
			})
			_, _, _ = refillSvc.GetContainerMovements(ctx, rc.ID.String(), 1, 10)
		}
	}

	// 3. ConsignmentService
	consignSvc := NewConsignmentService(db, consignmentRepo, transactionRepo)
	consignorPhone := "08123456789"
	consignor, _ := consignSvc.CreateConsignor(ctx, domain.CreateConsignorInput{
		Name:  "Pak RT Titip Roti",
		Phone: &consignorPhone,
	})
	if consignor != nil {
		_, _ = consignSvc.ListConsignors(ctx)
		newName := "Pak RT Titip Roti Updated"
		_, _ = consignSvc.UpdateConsignor(ctx, consignor.ID, domain.UpdateConsignorInput{
			Name: &newName,
		})
		_, _ = consignSvc.GenerateSettlement(ctx, consignor.ID, "Tester")
		_ = consignSvc.DeleteConsignor(ctx, consignor.ID)
	}

	// 4. NotificationService
	qClient := queue.NewClient("127.0.0.1:6381", "")
	defer qClient.Close()
	notifSvc := NewNotificationService(notifRepo, nil, qClient)
	dummyUID := uuid.New()
	_, _ = notifSvc.GetUserNotifications(ctx, dummyUID, 10, 0)
	_ = notifSvc.MarkAsRead(ctx, uuid.New())
	_ = notifSvc.MarkAllAsRead(ctx, dummyUID)
	_ = notifSvc.EnqueueLowStock(uuid.New().String(), "Kopi Sachet", 1, 5)

	lowStockPayload, _ := json.Marshal(queue.PayloadLowStock{
		ProductID:    uuid.New().String(),
		ProductName:  "Kopi Sachet",
		CurrentStock: 1,
		MinStock:     5,
	})
	_ = notifSvc.HandleLowStockTask(ctx, asynq.NewTask("notification:low_stock", lowStockPayload))

	newTxPayload, _ := json.Marshal(queue.PayloadNewTransaction{
		TransactionID: uuid.New().String(),
		Amount:        50000,
		CashierName:   "Budi",
	})
	_ = notifSvc.HandleNewTransactionTask(ctx, asynq.NewTask("notification:new_transaction", newTxPayload))

	// 5. StockOpnameService
	soSvc := NewStockOpnameService(db, stockOpnameRepo, productRepo, inventoryRepo)
	notes := "Sesi Opname Unit Test"
	soSession, err := soSvc.StartSession(ctx, domain.StartOpnameInput{
		Notes:     &notes,
		CreatedBy: "Tester",
	})
	if err == nil && soSession != nil {
		_, _, _ = soSvc.ListSessions(ctx, nil, 1, 10)
		_, _ = soSvc.GetSession(ctx, soSession.ID)
		if tp != nil {
			itemNotes := "Hitung fisik"
			_, _ = soSvc.RecordCount(ctx, domain.RecordCountInput{
				SessionID:     soSession.ID,
				ProductID:     tp.ID,
				PhysicalStock: 18,
				Notes:         &itemNotes,
				CountedBy:     "Tester",
			})
			_, _ = soSvc.GetVarianceReport(ctx, soSession.ID)
		}
		_, _ = soSvc.FinalizeSession(ctx, domain.FinalizeOpnameInput{
			SessionID:        soSession.ID,
			CompletedBy:      "Tester",
			ApplyAdjustments: true,
		})
	}
	_, _ = soSvc.GetShoppingList(ctx)
	_, _ = soSvc.GetNearExpiryReport(ctx, 30)

	// Cancel session test
	cancelSession, err := soSvc.StartSession(ctx, domain.StartOpnameInput{
		Notes:     &notes,
		CreatedBy: "Tester",
	})
	if err == nil && cancelSession != nil {
		_ = soSvc.CancelSession(ctx, cancelSession.ID)
	}

	// 6. POSService
	posSvc := NewPOSService(db, posRepo, productRepo, transactionRepo, inventoryRepo)
	if tp != nil {
		cartNotes := "Keranjang Tertunda Unit Test"
		cart, err := posSvc.HoldCart(ctx, domain.HoldCartInput{
			Items:  []domain.HoldCartItemInput{{ProductID: tp.ID, Quantity: 1}},
			Notes:  &cartNotes,
			HeldBy: "Tester",
		})
		if err == nil && cart != nil {
			_, _ = posSvc.ListHeldCarts(ctx)
			_, _ = posSvc.GetHeldCart(ctx, cart.ID)
			_, _ = posSvc.ResumeCart(ctx, cart.ID, "Tester")
		}

		// Discard
		cart2Notes := "Keranjang Batal"
		cart2, _ := posSvc.HoldCart(ctx, domain.HoldCartInput{
			Items:  []domain.HoldCartItemInput{{ProductID: tp.ID, Quantity: 1}},
			Notes:  &cart2Notes,
			HeldBy: "Tester",
		})
		if cart2 != nil {
			_ = posSvc.DiscardCart(ctx, cart2.ID)
		}
	}

	// 7. TransactionService
	txSvc := NewTransactionService(
		db,
		transactionRepo,
		productRepo,
		customerRepo,
		kasbonRepo,
		inventoryRepo,
		refillableRepo,
		cashFlowRepo,
		notifSvc,
	)

	if tp != nil {
		calc, _ := txSvc.CalculateCart(ctx, domain.CartCalculateInput{
			Items: []domain.CartItem{{ProductID: tp.ID, Quantity: 2}},
		})
		if calc != nil {
			txCash, err := txSvc.CreateTransaction(ctx, domain.TransactionCreateInput{
				Items:         []domain.TransactionItemInput{{ProductID: tp.ID, Quantity: 1}},
				PaymentMethod: domain.PaymentMethodCash,
				AmountPaid:    50000,
			})
			if err == nil && txCash != nil {
				if len(txCash.Items) > 0 {
					_, _ = posSvc.CreateRefund(ctx, domain.CreateRefundInput{
						TransactionID: txCash.ID,
						RefundMethod:  "cash",
						Reason:        "Barang cacat",
						RequestedBy:   "Tester",
						Items: []domain.RefundItemInput{
							{
								TransactionItemID: txCash.Items[0].ID,
								Quantity:          1,
							},
						},
					})
				}
				_ = txSvc.CancelTransaction(ctx, txCash.ID)
			}
		}

		// Kasbon Transaction
		var creditLimit int64 = 500000
		cust, _ := customerRepo.Create(ctx, domain.CustomerCreateInput{
			Name:        "Pelanggan Svc Test",
			CreditLimit: &creditLimit,
		})
		if cust != nil {
			_, _ = txSvc.CreateTransaction(ctx, domain.TransactionCreateInput{
				CustomerID:    &cust.ID,
				Items:         []domain.TransactionItemInput{{ProductID: tp.ID, Quantity: 1}},
				PaymentMethod: domain.PaymentMethodKasbon,
				AmountPaid:    0,
			})
		}
	}

	// 8. PaymentService
	mockPayServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"token": "snap-mock",
			"redirect_url": "https://midtrans",
			"status_code": "200",
			"transaction_status": "settlement",
			"actions": [{"name": "generate-qr-code", "url": "https://qr"}]
		}`))
	}))
	defer mockPayServer.Close()

	midtransCfg := &config.MidtransConfig{Environment: "sandbox", ServerKey: "SB-Mid-server-123"}
	paySvc := NewPaymentService(db, paymentRepo, transactionRepo, midtransCfg)
	paySvc.httpClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			mReq, _ := http.NewRequest(req.Method, mockPayServer.URL, req.Body)
			return http.DefaultTransport.RoundTrip(mReq)
		}),
	}

	_ = paySvc.HandleNotification(ctx, domain.MidtransNotification{
		OrderID:           "MOCK-ORDER-SVC",
		TransactionStatus: "settlement",
	})
	_, _ = paySvc.GetPaymentByTransactionID(ctx, uuid.New())
	_ = paySvc.ManualVerify(ctx, uuid.New(), "Tester")

	dummyTxID := uuid.New()
	if tp != nil {
		dummyTx, errCreateTx := txSvc.CreateTransaction(ctx, domain.TransactionCreateInput{
			Items:         []domain.TransactionItemInput{{ProductID: tp.ID, Quantity: 1}},
			PaymentMethod: domain.PaymentMethodCash,
			AmountPaid:    20000,
		})
		if errCreateTx != nil {
			t.Errorf("ERROR creating dummyTx: %v", errCreateTx)
		}
		if dummyTx != nil {
			dummyTxID = dummyTx.ID
		}
	}
	snapRes, errSnap := paySvc.GenerateSnapToken(ctx, domain.SnapTokenRequest{
		TransactionID: dummyTxID,
	})
	if errSnap != nil {
		t.Errorf("DEBUG errSnap: %v", errSnap)
	}
	if snapRes != nil {
		_ = snapRes.Token
	}

	qrisRes, errQRIS := paySvc.CreateQRISCharge(ctx, domain.QRISChargeRequest{
		TransactionID: dummyTxID,
	})
	if errQRIS != nil {
		t.Errorf("DEBUG errQRIS: %v", errQRIS)
	}
	if qrisRes != nil {
		_ = qrisRes.QRCodeURL
	}

	// Check status on created record
	pRec, _ := paymentRepo.GetByTransactionID(ctx, dummyTxID)
	if pRec != nil {
		_, _ = paySvc.CheckPaymentStatus(ctx, pRec.ID)

		orderSigData := pRec.OrderID + "200" + fmt.Sprintf("%d", pRec.GrossAmount) + midtransCfg.ServerKey
		oHash := sha512.Sum512([]byte(orderSigData))
		oSig := hex.EncodeToString(oHash[:])
		_ = paySvc.HandleNotification(ctx, domain.MidtransNotification{
			OrderID:           pRec.OrderID,
			StatusCode:        "200",
			GrossAmount:       fmt.Sprintf("%d", pRec.GrossAmount),
			TransactionStatus: "settlement",
			SignatureKey:      oSig,
		})
		_ = paySvc.HandleNotification(ctx, domain.MidtransNotification{
			OrderID:           pRec.OrderID,
			StatusCode:        "200",
			GrossAmount:       fmt.Sprintf("%d", pRec.GrossAmount),
			TransactionStatus: "capture",
			FraudStatus:       "accept",
			SignatureKey:      oSig,
		})
		_ = paySvc.HandleNotification(ctx, domain.MidtransNotification{
			OrderID:           pRec.OrderID,
			StatusCode:        "200",
			GrossAmount:       fmt.Sprintf("%d", pRec.GrossAmount),
			TransactionStatus: "cancel",
			SignatureKey:      oSig,
		})
		_ = paySvc.HandleNotification(ctx, domain.MidtransNotification{
			OrderID:           pRec.OrderID,
			StatusCode:        "200",
			GrossAmount:       fmt.Sprintf("%d", pRec.GrossAmount),
			TransactionStatus: "expire",
			SignatureKey:      oSig,
		})

		_ = paySvc.ManualVerify(ctx, pRec.ID, "Tester")
		_ = paySvc.ManualVerify(ctx, pRec.ID, "Tester") // already verified branch
	}

	// Also test prod config URLs
	prodPaySvc := NewPaymentService(db, paymentRepo, transactionRepo, &config.MidtransConfig{Environment: "production"})
	prodPaySvc.httpClient = paySvc.httpClient
	_, _ = prodPaySvc.GenerateSnapToken(ctx, domain.SnapTokenRequest{
		TransactionID: dummyTxID,
	})

	// Test URLs and signature verification
	_ = paySvc.getSnapURL()
	_ = paySvc.getCoreAPIURL()
	_ = prodPaySvc.getSnapURL()
	_ = prodPaySvc.getCoreAPIURL()

	orderID := "ORDER-123"
	statusCode := "200"
	grossAmount := "50000.00"
	serverKey := midtransCfg.ServerKey
	sigData := orderID + statusCode + grossAmount + serverKey
	hash := sha512.Sum512([]byte(sigData))
	sig := hex.EncodeToString(hash[:])
	_ = paySvc.verifySignature(domain.MidtransNotification{
		OrderID:      orderID,
		StatusCode:   statusCode,
		GrossAmount:  grossAmount,
		SignatureKey: sig,
	})

	// Service error branches testing
	categoryRepo := repository.NewCategoryRepository(db)
	catSvc := NewCategoryService(categoryRepo)
	nonID := uuid.New()
	_, _ = catSvc.Get(ctx, nonID)
	_, _ = catSvc.Update(ctx, nonID, domain.CategoryUpdateInput{})
	_ = catSvc.Delete(ctx, nonID)

	_, _ = consignSvc.UpdateConsignor(ctx, nonID, domain.UpdateConsignorInput{})
	_ = consignSvc.DeleteConsignor(ctx, nonID)

	_, _ = soSvc.GetSession(ctx, nonID)
	_, _ = soSvc.GetVarianceReport(ctx, nonID)
	_ = soSvc.CancelSession(ctx, nonID)

	_, _ = posSvc.GetHeldCart(ctx, nonID)
	_, _ = posSvc.ResumeCart(ctx, nonID, "Tester")
	_ = posSvc.DiscardCart(ctx, nonID)
	_, _ = posSvc.HoldCart(ctx, domain.HoldCartInput{Items: nil})
	_, _ = posSvc.CreateRefund(ctx, domain.CreateRefundInput{TransactionID: nonID})

	_, _ = txSvc.CreateTransaction(ctx, domain.TransactionCreateInput{Items: nil})
	_ = txSvc.CancelTransaction(ctx, nonID)

	_, _ = cfSvc.CloseDrawer(ctx, domain.CloseDrawerInput{SessionID: nonID})

	// Close drawer at end
	if session != nil {
		closeNotes := "Selesai test"
		_, _ = cfSvc.CloseDrawer(ctx, domain.CloseDrawerInput{
			SessionID:      session.ID,
			ClosingBalance: 140000,
			ClosedBy:       "Tester",
			Notes:          &closeNotes,
		})
	}
}
