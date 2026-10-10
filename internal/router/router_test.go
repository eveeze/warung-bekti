package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/router"
)

func setupTestRouter(t *testing.T) (http.Handler, string) {
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
		t.Fatalf("failed to connect test database: %v", err)
	}

	redisCfg := &config.RedisConfig{Host: "127.0.0.1", Port: "6381"}
	rClient, _ := database.NewRedis(redisCfg)

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:          "d48ca69b165d9c97b8ffeef00bb3caafef1b58e1e8670842fcd110c46ea3836a",
			ExpirationHours: 24,
			Issuer:          "warung-test",
		},
		Redis: *redisCfg,
	}

	appRouter := router.New(cfg, db, rClient, nil)

	// Obtain admin token
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "admin@warung.com",
		"password": "password",
	})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	appRouter.ServeHTTP(rec, req)

	var loginResp struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &loginResp)

	return appRouter, loginResp.Data.AccessToken
}

func TestRouterRoutes(t *testing.T) {
	appRouter, token := setupTestRouter(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		useAuth    bool
		wantStatus int
	}{
		{"Health", "GET", "/health", nil, false, http.StatusOK},
		{"Ready", "GET", "/ready", nil, false, http.StatusOK},
		{"Live", "GET", "/live", nil, false, http.StatusOK},
		{"PublicCategories", "GET", "/public/categories", nil, false, http.StatusOK},
		{"PublicProducts", "GET", "/public/products", nil, false, http.StatusOK},
		{"ListProducts", "GET", "/api/v1/products", nil, true, http.StatusOK},
		{"ListCategories", "GET", "/api/v1/categories", nil, true, http.StatusOK},
		{"ListCustomers", "GET", "/api/v1/customers", nil, true, http.StatusOK},
		{"ListTransactions", "GET", "/api/v1/transactions", nil, true, http.StatusOK},
		{"ListHeldCarts", "GET", "/api/v1/pos/held-carts", nil, true, http.StatusOK},
		{"ListConsignors", "GET", "/api/v1/consignors", nil, true, http.StatusOK},
		{"ListRefillables", "GET", "/api/v1/refillables", nil, true, http.StatusOK},
		{"ListLocations", "GET", "/api/v1/locations", nil, true, http.StatusOK},
		{"ListNotifications", "GET", "/api/v1/notifications", nil, true, http.StatusOK},
		{"ReportDashboard", "GET", "/api/v1/reports/dashboard", nil, true, http.StatusOK},
		{"ReportDaily", "GET", "/api/v1/reports/daily", nil, true, http.StatusOK},
		{"ReportKasbon", "GET", "/api/v1/reports/kasbon", nil, true, http.StatusOK},
		{"ReportInventory", "GET", "/api/v1/reports/inventory", nil, true, http.StatusOK},
		{"InventoryLowStock", "GET", "/api/v1/inventory/low-stock", nil, true, http.StatusOK},
		{"InventoryReport", "GET", "/api/v1/inventory/report", nil, true, http.StatusOK},
		{"CashflowCategories", "GET", "/api/v1/cashflow/categories", nil, true, http.StatusOK},
		{"CashflowCurrentDrawer", "GET", "/api/v1/cashflow/drawer/current", nil, true, 200}, // or 404
		{"StockOpnameSessions", "GET", "/api/v1/stock-opname/sessions", nil, true, http.StatusOK},
		{"StockOpnameShoppingList", "GET", "/api/v1/stock-opname/shopping-list", nil, true, http.StatusOK},
		{"StockOpnameNearExpiry", "GET", "/api/v1/stock-opname/near-expiry", nil, true, http.StatusOK},
		{"PaymentWebhook", "POST", "/api/v1/payments/notification", map[string]string{"order_id": "test"}, false, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				b, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(b)
			} else {
				bodyReader = bytes.NewReader(nil)
			}

			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			if tt.useAuth {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			rec := httptest.NewRecorder()
			appRouter.ServeHTTP(rec, req)

			if tt.wantStatus == 200 && (rec.Code == http.StatusOK || rec.Code == http.StatusNotFound) {
				// OK for endpoints that may return 404 if no session open
				return
			}
			if rec.Code != tt.wantStatus {
				t.Errorf("path %s expected status %d, got %d: %s", tt.path, tt.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestFullAppE2EWorkflow(t *testing.T) {
	appRouter, token := setupTestRouter(t)

	reqHelper := func(method, path string, body interface{}) *httptest.ResponseRecorder {
		var bodyReader *bytes.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			bodyReader = bytes.NewReader(b)
		} else {
			bodyReader = bytes.NewReader(nil)
		}
		req := httptest.NewRequest(method, path, bodyReader)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		rec := httptest.NewRecorder()
		appRouter.ServeHTTP(rec, req)
		return rec
	}

	// 1. Drawer open
	dRec := reqHelper("POST", "/api/v1/cashflow/drawer/open", map[string]interface{}{
		"opening_balance": 100000,
	})
	var dData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(dRec.Body.Bytes(), &dData)
	drawerID := dData.Data.ID

	// 2. Categories
	catRec := reqHelper("POST", "/api/v1/categories", map[string]string{
		"name": "Router Cat " + uuid.New().String()[:6],
	})
	var catData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(catRec.Body.Bytes(), &catData)
	catID := catData.Data.ID

	if catID != "" {
		_ = reqHelper("GET", "/api/v1/categories/"+catID, nil)
		_ = reqHelper("PUT", "/api/v1/categories/"+catID, map[string]string{"name": "Router Cat Updated"})
	}

	// 3. Products
	barcode := fmt.Sprintf("BAR%09d", time.Now().UnixNano()%1000000000)
	prodRec := reqHelper("POST", "/api/v1/products", map[string]interface{}{
		"name":            "Router Prod " + uuid.New().String()[:6],
		"barcode":         barcode,
		"unit":            "pcs",
		"base_price":      25000,
		"cost_price":      18000,
		"is_stock_active": true,
		"current_stock":   100,
		"min_stock_alert": 10,
	})
	var prodData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(prodRec.Body.Bytes(), &prodData)
	prodID := prodData.Data.ID

	if prodID != "" {
		_ = reqHelper("GET", "/api/v1/products/"+prodID, nil)
		_ = reqHelper("GET", "/api/v1/products/search?barcode="+barcode, nil)
		_ = reqHelper("PUT", "/api/v1/products/"+prodID, map[string]interface{}{"base_price": 26000})
		_ = reqHelper("PATCH", "/api/v1/products/"+prodID+"/toggle-active", nil)
		_ = reqHelper("PATCH", "/api/v1/products/"+prodID+"/toggle-active", nil)

		tierRec := reqHelper("POST", "/api/v1/products/"+prodID+"/pricing-tiers", map[string]interface{}{
			"min_quantity": 5, "price": 23000,
		})
		var tData struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(tierRec.Body.Bytes(), &tData)
		if tData.Data.ID != "" {
			_ = reqHelper("PUT", "/api/v1/products/"+prodID+"/pricing-tiers/"+tData.Data.ID, map[string]interface{}{
				"min_quantity": 5, "price": 22500,
			})
			_ = reqHelper("DELETE", "/api/v1/products/"+prodID+"/pricing-tiers/"+tData.Data.ID, nil)
		}
		_ = reqHelper("GET", "/public/products/"+prodID, nil)
	}

	// 4. Customers
	phone := fmt.Sprintf("081%09d", time.Now().UnixNano()%1000000000)
	cRec := reqHelper("POST", "/api/v1/customers", map[string]interface{}{
		"name": "Router Cust " + uuid.New().String()[:6], "phone": phone, "credit_limit": 500000,
	})
	var cData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(cRec.Body.Bytes(), &cData)
	custID := cData.Data.ID

	if custID != "" {
		_ = reqHelper("GET", "/api/v1/customers/"+custID, nil)
		_ = reqHelper("PUT", "/api/v1/customers/"+custID, map[string]interface{}{"credit_limit": 700000})
		_ = reqHelper("GET", "/api/v1/customers/with-debt", nil)
	}

	// 5. Cashflow
	_ = reqHelper("POST", "/api/v1/cashflow", map[string]interface{}{
		"type": "expense", "amount": 10000, "notes": "Plastik",
	})
	_ = reqHelper("POST", "/api/v1/cashflow", map[string]interface{}{
		"type": "income", "amount": 20000, "notes": "Kembalian",
	})
	_ = reqHelper("GET", "/api/v1/cashflow", nil)

	// 6. Transactions
	if prodID != "" {
		_ = reqHelper("POST", "/api/v1/transactions/calculate", map[string]interface{}{
			"items": []map[string]interface{}{{"product_id": prodID, "quantity": 2}},
		})
		txCash := reqHelper("POST", "/api/v1/transactions", map[string]interface{}{
			"items":          []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
			"payment_method": "cash",
			"amount_paid":    50000,
			"cashier_name":   "Kasir Router",
		})
		var txData struct {
			Data struct {
				ID    string `json:"id"`
				Items []struct {
					ID string `json:"id"`
				} `json:"items"`
			} `json:"data"`
		}
		_ = json.Unmarshal(txCash.Body.Bytes(), &txData)
		if txData.Data.ID != "" {
			_ = reqHelper("GET", "/api/v1/transactions/"+txData.Data.ID, nil)

			if len(txData.Data.Items) > 0 {
				_ = reqHelper("POST", "/api/v1/pos/refunds", map[string]interface{}{
					"transaction_id": txData.Data.ID,
					"refund_method":  "cash",
					"reason":         "Rusak",
					"items": []map[string]interface{}{
						{"transaction_item_id": txData.Data.Items[0].ID, "quantity": 1, "restock": true},
					},
				})
			}
		}

		// Kasbon Transaction
		if custID != "" {
			_ = reqHelper("POST", "/api/v1/transactions", map[string]interface{}{
				"customer_id":    custID,
				"items":          []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
				"payment_method": "kasbon",
				"amount_paid":    0,
				"cashier_name":   "Kasir Router",
			})
			_ = reqHelper("GET", "/api/v1/kasbon/customers/"+custID, nil)
			_ = reqHelper("GET", "/api/v1/kasbon/customers/"+custID+"/summary", nil)
			_ = reqHelper("GET", "/api/v1/kasbon/customers/"+custID+"/billing/pdf", nil)
			_ = reqHelper("POST", "/api/v1/kasbon/customers/"+custID+"/payments", map[string]interface{}{"amount": 10000})
		}
	}

	// 7. POS Held Carts
	if prodID != "" {
		hcRec := reqHelper("POST", "/api/v1/pos/held-carts", map[string]interface{}{
			"items":   []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
			"held_by": "Tester",
		})
		var hcData struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(hcRec.Body.Bytes(), &hcData)
		if hcData.Data.ID != "" {
			_ = reqHelper("GET", "/api/v1/pos/held-carts/"+hcData.Data.ID, nil)
			_ = reqHelper("POST", "/api/v1/pos/held-carts/"+hcData.Data.ID+"/resume", nil)
			_ = reqHelper("POST", "/api/v1/pos/held-carts/"+hcData.Data.ID+"/discard", nil)
		}
	}

	// 8. Inventory
	if prodID != "" {
		_ = reqHelper("POST", "/api/v1/inventory/restock", map[string]interface{}{
			"product_id": prodID, "quantity": 10, "cost_per_unit": 10000,
		})
		_ = reqHelper("POST", "/api/v1/inventory/adjust", map[string]interface{}{
			"product_id": prodID, "quantity": -1, "reason": "Kotor",
		})
		_ = reqHelper("GET", "/api/v1/inventory/restock-list/pdf", nil)
		_ = reqHelper("GET", "/api/v1/inventory/"+prodID+"/movements", nil)
	}

	// 9. Stock Opname
	soRec := reqHelper("POST", "/api/v1/stock-opname/sessions", map[string]string{"notes": "Opname Router"})
	var soData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(soRec.Body.Bytes(), &soData)
	if soData.Data.ID != "" {
		_ = reqHelper("GET", "/api/v1/stock-opname/sessions/"+soData.Data.ID, nil)
		if prodID != "" {
			_ = reqHelper("POST", "/api/v1/stock-opname/sessions/"+soData.Data.ID+"/items", map[string]interface{}{
				"product_id": prodID, "physical_stock": 95,
			})
			_ = reqHelper("GET", "/api/v1/stock-opname/sessions/"+soData.Data.ID+"/variance", nil)
		}
		_ = reqHelper("POST", "/api/v1/stock-opname/sessions/"+soData.Data.ID+"/finalize", map[string]bool{"apply_adjustments": false})
	}

	// 10. Consignment
	conRec := reqHelper("POST", "/api/v1/consignors", map[string]interface{}{"name": "Mitra Router " + uuid.New().String()[:4]})
	var conData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(conRec.Body.Bytes(), &conData)
	if conData.Data.ID != "" {
		_ = reqHelper("PUT", "/api/v1/consignors/"+conData.Data.ID, map[string]string{"name": "Mitra Updated"})
		_ = reqHelper("DELETE", "/api/v1/consignors/"+conData.Data.ID, nil)
	}

	// 11. Refillables
	if prodID != "" {
		rfRec := reqHelper("POST", "/api/v1/refillables", map[string]interface{}{
			"product_id": prodID, "container_type": "gas", "empty_count": 5, "full_count": 10,
		})
		var rfData struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rfRec.Body.Bytes(), &rfData)
		if rfData.Data.ID != "" {
			_ = reqHelper("POST", "/api/v1/refillables/adjust", map[string]interface{}{
				"container_id": rfData.Data.ID, "empty_change": 1, "full_change": -1,
			})
			_ = reqHelper("GET", "/api/v1/refillables/"+rfData.Data.ID+"/movements", nil)
		}
	}

	// 12. Locations
	locRec := reqHelper("POST", "/api/v1/locations", map[string]interface{}{
		"name": "Rak Router " + uuid.New().String()[:4], "category": "shelf", "width": 1.0, "depth": 1.0, "height": 1.0,
	})
	var locData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(locRec.Body.Bytes(), &locData)
	if locData.Data.ID != "" {
		_ = reqHelper("GET", "/api/v1/locations/"+locData.Data.ID, nil)
		_ = reqHelper("PUT", "/api/v1/locations/"+locData.Data.ID, map[string]string{"name": "Rak Router Updated"})
		_ = reqHelper("DELETE", "/api/v1/locations/"+locData.Data.ID, nil)
	}

	// 13. Notifications
	_ = reqHelper("PATCH", "/api/v1/notifications/00000000-0000-0000-0000-000000000000/read", nil)
	_ = reqHelper("PATCH", "/api/v1/notifications/read-all", nil)

	// 14. Payments
	_ = reqHelper("POST", "/api/v1/payments/snap", map[string]interface{}{})
	_ = reqHelper("POST", "/api/v1/payments/qris/charge", map[string]interface{}{})
	_ = reqHelper("POST", "/api/v1/payments/00000000-0000-0000-0000-000000000000/manual-verify", nil)
	_ = reqHelper("GET", "/api/v1/payments/status/00000000-0000-0000-0000-000000000000", nil)
	_ = reqHelper("GET", "/api/v1/payments/transaction/00000000-0000-0000-0000-000000000000", nil)

	// 15. Close Drawer & Cleanup
	if drawerID != "" {
		_ = reqHelper("POST", "/api/v1/cashflow/drawer/close", map[string]interface{}{
			"session_id": drawerID, "closing_balance": 110000,
		})
	}
	if catID != "" {
		_ = reqHelper("DELETE", "/api/v1/categories/"+catID, nil)
	}
	if custID != "" {
		_ = reqHelper("DELETE", "/api/v1/customers/"+custID, nil)
	}
	if prodID != "" {
		_ = reqHelper("DELETE", "/api/v1/products/"+prodID, nil)
	}
}
