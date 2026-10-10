package service_test

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

type e2eHelper struct {
	router         http.Handler
	adminToken     string
	cashierToken   string
	inventoryToken string
}

func setupE2E(t *testing.T) *e2eHelper {
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
		t.Fatalf("cannot connect db: %v", err)
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

	login := func(email, pass string) string {
		b, _ := json.Marshal(map[string]string{"email": email, "password": pass})
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		appRouter.ServeHTTP(rec, req)
		var res struct {
			Data struct {
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		return res.Data.AccessToken
	}

	return &e2eHelper{
		router:         appRouter,
		adminToken:     login("admin@warung.com", "password"),
		cashierToken:   login("cashier@warung.com", "password"),
		inventoryToken: login("inventory@warung.com", "password"),
	}
}

func (h *e2eHelper) request(token, method, path string, body interface{}) *httptest.ResponseRecorder {
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
	h.router.ServeHTTP(rec, req)
	return rec
}

func TestFullGoE2ESuite(t *testing.T) {
	h := setupE2E(t)

	// 1. Health
	rec := h.request("", "GET", "/health", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("health: %d", rec.Code)
	}
	_ = h.request("", "GET", "/ready", nil)
	_ = h.request("", "GET", "/live", nil)

	// 2. Auth & RBAC
	_ = h.request("", "POST", "/auth/login", map[string]string{"email": "admin@warung.com", "password": "wrong"})
	_ = h.request("", "POST", "/auth/register", map[string]string{
		"name": "Dev User", "email": "dev_" + uuid.New().String()[:8] + "@warung.com", "password": "password123", "role": "cashier",
	})

	// 3. Category & Public
	catRec := h.request(h.adminToken, "POST", "/api/v1/categories", map[string]string{"name": "Cat Go " + uuid.New().String()[:8]})
	var catData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(catRec.Body.Bytes(), &catData)
	catID := catData.Data.ID

	_ = h.request(h.adminToken, "GET", "/api/v1/categories", nil)
	if catID != "" {
		_ = h.request(h.adminToken, "GET", "/api/v1/categories/"+catID, nil)
		_ = h.request(h.adminToken, "PUT", "/api/v1/categories/"+catID, map[string]string{"name": "Cat Updated"})
	}
	_ = h.request("", "GET", "/public/categories", nil)
	_ = h.request("", "GET", "/public/products", nil)

	// 4. Products
	barcode := fmt.Sprintf("BAR%09d", time.Now().UnixNano()%1000000000)
	prodRec := h.request(h.adminToken, "POST", "/api/v1/products", map[string]interface{}{
		"name":            "Prod Go " + uuid.New().String()[:8],
		"barcode":         barcode,
		"unit":            "pcs",
		"base_price":      20000,
		"cost_price":      15000,
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
		_ = h.request(h.adminToken, "GET", "/api/v1/products", nil)
		_ = h.request(h.adminToken, "GET", "/api/v1/products/"+prodID, nil)
		_ = h.request(h.adminToken, "GET", "/api/v1/products/search?barcode="+barcode, nil)
		_ = h.request(h.adminToken, "GET", "/api/v1/products/low-stock", nil)
		_ = h.request(h.adminToken, "PUT", "/api/v1/products/"+prodID, map[string]interface{}{"base_price": 21000})
		_ = h.request(h.adminToken, "PATCH", "/api/v1/products/"+prodID+"/toggle-active", nil)
		_ = h.request(h.adminToken, "PATCH", "/api/v1/products/"+prodID+"/toggle-active", nil)

		tierRec := h.request(h.adminToken, "POST", "/api/v1/products/"+prodID+"/pricing-tiers", map[string]interface{}{
			"min_quantity": 5, "price": 19000,
		})
		var tData struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(tierRec.Body.Bytes(), &tData)
		if tData.Data.ID != "" {
			_ = h.request(h.adminToken, "PUT", "/api/v1/products/"+prodID+"/pricing-tiers/"+tData.Data.ID, map[string]interface{}{
				"min_quantity": 5, "price": 18500,
			})
			_ = h.request(h.adminToken, "DELETE", "/api/v1/products/"+prodID+"/pricing-tiers/"+tData.Data.ID, nil)
		}
		_ = h.request("", "GET", "/public/products/"+prodID, nil)
	}

	// 5. Drawer & Cashflow
	dRec := h.request(h.cashierToken, "POST", "/api/v1/cashflow/drawer/open", map[string]interface{}{
		"opening_balance": 100000,
	})
	var dData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(dRec.Body.Bytes(), &dData)
	drawerID := dData.Data.ID

	_ = h.request(h.cashierToken, "GET", "/api/v1/cashflow/categories", nil)
	_ = h.request(h.cashierToken, "GET", "/api/v1/cashflow/drawer/current", nil)
	_ = h.request(h.cashierToken, "POST", "/api/v1/cashflow", map[string]interface{}{
		"type": "expense", "amount": 10000, "notes": "Plastik",
	})
	_ = h.request(h.cashierToken, "POST", "/api/v1/cashflow", map[string]interface{}{
		"type": "income", "amount": 20000, "notes": "Kembalian",
	})
	_ = h.request(h.cashierToken, "GET", "/api/v1/cashflow", nil)

	// 6. Customers & Kasbon
	phone := fmt.Sprintf("081%09d", time.Now().UnixNano()%1000000000)
	cRec := h.request(h.cashierToken, "POST", "/api/v1/customers", map[string]interface{}{
		"name": "Cust Go " + uuid.New().String()[:8], "phone": phone, "credit_limit": 500000,
	})
	var cData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(cRec.Body.Bytes(), &cData)
	custID := cData.Data.ID

	if custID != "" {
		_ = h.request(h.cashierToken, "GET", "/api/v1/customers", nil)
		_ = h.request(h.cashierToken, "GET", "/api/v1/customers/"+custID, nil)
		_ = h.request(h.cashierToken, "PUT", "/api/v1/customers/"+custID, map[string]interface{}{"credit_limit": 600000})
		_ = h.request(h.cashierToken, "GET", "/api/v1/customers/with-debt", nil)
	}

	// 7. Transactions
	if prodID != "" {
		_ = h.request(h.cashierToken, "POST", "/api/v1/transactions/calculate", map[string]interface{}{
			"items": []map[string]interface{}{{"product_id": prodID, "quantity": 2}},
		})
		txCash := h.request(h.cashierToken, "POST", "/api/v1/transactions", map[string]interface{}{
			"items":          []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
			"payment_method": "cash",
			"amount_paid":    50000,
			"cashier_name":   "Kasir Go",
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
			_ = h.request(h.cashierToken, "GET", "/api/v1/transactions", nil)
			_ = h.request(h.cashierToken, "GET", "/api/v1/transactions/"+txData.Data.ID, nil)

			if len(txData.Data.Items) > 0 {
				_ = h.request(h.cashierToken, "POST", "/api/v1/pos/refunds", map[string]interface{}{
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
			_ = h.request(h.cashierToken, "POST", "/api/v1/transactions", map[string]interface{}{
				"customer_id":    custID,
				"items":          []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
				"payment_method": "kasbon",
				"amount_paid":    0,
				"cashier_name":   "Kasir Go",
			})
			_ = h.request(h.cashierToken, "GET", "/api/v1/kasbon/customers/"+custID, nil)
			_ = h.request(h.cashierToken, "GET", "/api/v1/kasbon/customers/"+custID+"/summary", nil)
			_ = h.request(h.cashierToken, "GET", "/api/v1/kasbon/customers/"+custID+"/billing/pdf", nil)
			_ = h.request(h.cashierToken, "POST", "/api/v1/kasbon/customers/"+custID+"/payments", map[string]interface{}{"amount": 10000})
		}
	}

	// 8. POS Held Carts
	if prodID != "" {
		hcRec := h.request(h.cashierToken, "POST", "/api/v1/pos/held-carts", map[string]interface{}{
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
			_ = h.request(h.cashierToken, "GET", "/api/v1/pos/held-carts", nil)
			_ = h.request(h.cashierToken, "GET", "/api/v1/pos/held-carts/"+hcData.Data.ID, nil)
			_ = h.request(h.cashierToken, "POST", "/api/v1/pos/held-carts/"+hcData.Data.ID+"/resume", nil)
			_ = h.request(h.cashierToken, "POST", "/api/v1/pos/held-carts/"+hcData.Data.ID+"/discard", nil)
		}
	}

	// 9. Inventory
	if prodID != "" {
		_ = h.request(h.inventoryToken, "POST", "/api/v1/inventory/restock", map[string]interface{}{
			"product_id": prodID, "quantity": 10, "cost_per_unit": 10000,
		})
		_ = h.request(h.adminToken, "POST", "/api/v1/inventory/adjust", map[string]interface{}{
			"product_id": prodID, "quantity": -1, "reason": "Kotor",
		})
		_ = h.request(h.inventoryToken, "GET", "/api/v1/inventory/low-stock", nil)
		_ = h.request(h.inventoryToken, "GET", "/api/v1/inventory/report", nil)
		_ = h.request(h.inventoryToken, "GET", "/api/v1/inventory/restock-list/pdf", nil)
		_ = h.request(h.inventoryToken, "GET", "/api/v1/inventory/"+prodID+"/movements", nil)
	}

	// 10. Stock Opname
	soRec := h.request(h.inventoryToken, "POST", "/api/v1/stock-opname/sessions", map[string]string{"notes": "Opname Go"})
	var soData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(soRec.Body.Bytes(), &soData)
	if soData.Data.ID != "" {
		_ = h.request(h.inventoryToken, "GET", "/api/v1/stock-opname/sessions", nil)
		_ = h.request(h.inventoryToken, "GET", "/api/v1/stock-opname/sessions/"+soData.Data.ID, nil)
		if prodID != "" {
			_ = h.request(h.inventoryToken, "POST", "/api/v1/stock-opname/sessions/"+soData.Data.ID+"/items", map[string]interface{}{
				"product_id": prodID, "physical_stock": 95,
			})
			_ = h.request(h.inventoryToken, "GET", "/api/v1/stock-opname/sessions/"+soData.Data.ID+"/variance", nil)
		}
		_ = h.request(h.inventoryToken, "POST", "/api/v1/stock-opname/sessions/"+soData.Data.ID+"/finalize", map[string]bool{"apply_adjustments": false})
	}
	_ = h.request(h.inventoryToken, "GET", "/api/v1/stock-opname/shopping-list", nil)
	_ = h.request(h.inventoryToken, "GET", "/api/v1/stock-opname/near-expiry", nil)

	// 11. Consignment
	conRec := h.request(h.adminToken, "POST", "/api/v1/consignors", map[string]interface{}{"name": "Mitra Go " + uuid.New().String()[:4]})
	var conData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(conRec.Body.Bytes(), &conData)
	if conData.Data.ID != "" {
		_ = h.request(h.adminToken, "GET", "/api/v1/consignors", nil)
		_ = h.request(h.adminToken, "PUT", "/api/v1/consignors/"+conData.Data.ID, map[string]string{"name": "Mitra Updated"})
		_ = h.request(h.adminToken, "DELETE", "/api/v1/consignors/"+conData.Data.ID, nil)
	}

	// 12. Refillables
	if prodID != "" {
		rfRec := h.request(h.inventoryToken, "POST", "/api/v1/refillables", map[string]interface{}{
			"product_id": prodID, "container_type": "gas", "empty_count": 5, "full_count": 10,
		})
		var rfData struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rfRec.Body.Bytes(), &rfData)
		if rfData.Data.ID != "" {
			_ = h.request(h.inventoryToken, "GET", "/api/v1/refillables", nil)
			_ = h.request(h.inventoryToken, "POST", "/api/v1/refillables/adjust", map[string]interface{}{
				"container_id": rfData.Data.ID, "empty_change": 1, "full_change": -1,
			})
			_ = h.request(h.inventoryToken, "GET", "/api/v1/refillables/"+rfData.Data.ID+"/movements", nil)
		}
	}

	// 13. Locations
	locRec := h.request(h.adminToken, "POST", "/api/v1/locations", map[string]interface{}{
		"name": "Rak Go " + uuid.New().String()[:4], "category": "shelf", "width": 1.0, "depth": 1.0, "height": 1.0,
	})
	var locData struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(locRec.Body.Bytes(), &locData)
	if locData.Data.ID != "" {
		_ = h.request(h.adminToken, "GET", "/api/v1/locations", nil)
		_ = h.request(h.adminToken, "GET", "/api/v1/locations/"+locData.Data.ID, nil)
		_ = h.request(h.adminToken, "PUT", "/api/v1/locations/"+locData.Data.ID, map[string]string{"name": "Rak Go Updated"})
		_ = h.request(h.adminToken, "DELETE", "/api/v1/locations/"+locData.Data.ID, nil)
	}

	// 14. Reports
	_ = h.request(h.adminToken, "GET", "/api/v1/reports/daily", nil)
	_ = h.request(h.adminToken, "GET", "/api/v1/reports/kasbon", nil)
	_ = h.request(h.adminToken, "GET", "/api/v1/reports/inventory", nil)
	_ = h.request(h.adminToken, "GET", "/api/v1/reports/dashboard", nil)

	// 15. Notifications
	_ = h.request(h.adminToken, "GET", "/api/v1/notifications", nil)
	_ = h.request(h.adminToken, "PATCH", "/api/v1/notifications/00000000-0000-0000-0000-000000000000/read", nil)
	_ = h.request(h.adminToken, "PATCH", "/api/v1/notifications/read-all", nil)

	// 16. Payments
	_ = h.request("", "POST", "/api/v1/payments/notification", map[string]string{"order_id": "MOCK"})
	_ = h.request(h.cashierToken, "POST", "/api/v1/payments/snap", map[string]interface{}{})
	_ = h.request(h.cashierToken, "POST", "/api/v1/payments/qris/charge", map[string]interface{}{})
	_ = h.request(h.adminToken, "POST", "/api/v1/payments/00000000-0000-0000-0000-000000000000/manual-verify", nil)
	_ = h.request(h.cashierToken, "GET", "/api/v1/payments/status/00000000-0000-0000-0000-000000000000", nil)
	_ = h.request(h.cashierToken, "GET", "/api/v1/payments/transaction/00000000-0000-0000-0000-000000000000", nil)

	// 17. Close Drawer
	if drawerID != "" {
		_ = h.request(h.cashierToken, "POST", "/api/v1/cashflow/drawer/close", map[string]interface{}{
			"session_id": drawerID, "closing_balance": 110000,
		})
	}

	// 18. Cleanup
	if catID != "" {
		_ = h.request(h.adminToken, "DELETE", "/api/v1/categories/"+catID, nil)
	}
	if custID != "" {
		_ = h.request(h.adminToken, "DELETE", "/api/v1/customers/"+custID, nil)
	}
	if prodID != "" {
		_ = h.request(h.adminToken, "DELETE", "/api/v1/products/"+prodID, nil)
	}

	// 19. RBAC Rejections
	unauth := h.request("", "GET", "/api/v1/products", nil)
	if unauth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 unauth, got %d", unauth.Code)
	}
	forbid := h.request(h.cashierToken, "POST", "/api/v1/users", map[string]string{"name": "Hacker"})
	if forbid.Code != http.StatusForbidden {
		t.Errorf("expected 403 forbidden, got %d", forbid.Code)
	}
}
