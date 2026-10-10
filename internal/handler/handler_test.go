package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/database"
	"github.com/eveeze/warung-backend/internal/handler"
	"github.com/eveeze/warung-backend/internal/repository"
	"github.com/eveeze/warung-backend/internal/router"
	"github.com/eveeze/warung-backend/internal/service"
)

func setupTestEnvironment(t *testing.T) (http.Handler, string, string, *database.PostgresDB) {
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
		t.Fatalf("failed to connect live db: %v", err)
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

	// Admin login
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
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &loginResp)

	return appRouter, loginResp.Data.AccessToken, loginResp.Data.RefreshToken, db
}

func doRequest(handler http.Handler, token, method, path string, body interface{}) *httptest.ResponseRecorder {
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
	handler.ServeHTTP(rec, req)
	return rec
}

func TestAllHandlersThroughMux(t *testing.T) {
	appRouter, token, refreshToken, _ := setupTestEnvironment(t)

	// 1. Health Handler
	rec := doRequest(appRouter, "", "GET", "/health", nil)
	if rec.Code != http.StatusOK {
		t.Errorf("health failed: %d", rec.Code)
	}
	_ = doRequest(appRouter, "", "GET", "/ready", nil)
	_ = doRequest(appRouter, "", "GET", "/live", nil)

	// 2. Auth Handler
	ts := time.Now().UnixNano()
	registerPayload := map[string]interface{}{
		"name":     "Handler Tester",
		"email":    "handler_test_" + uuid.New().String()[:8] + "@warung.com",
		"password": "password123",
		"role":     "cashier",
	}
	rec = doRequest(appRouter, "", "POST", "/auth/register", registerPayload)
	if rec.Code != http.StatusCreated {
		t.Errorf("register failed: %d: %s", rec.Code, rec.Body.String())
	}
	_ = doRequest(appRouter, "", "POST", "/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	})

	// 3. User Handler
	userPayload := map[string]interface{}{
		"name":     "Staff Created",
		"email":    "staff_" + uuid.New().String()[:8] + "@warung.com",
		"password": "password123",
		"role":     "cashier",
	}
	rec = doRequest(appRouter, token, "POST", "/api/v1/users", userPayload)
	if rec.Code != http.StatusCreated {
		t.Errorf("create user failed: %d: %s", rec.Code, rec.Body.String())
	}
	var userResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &userResp)
	createdUID := userResp.Data.ID

	if createdUID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/users?page=1&per_page=5&search=Staff", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/users/"+createdUID, nil)
		_ = doRequest(appRouter, token, "PUT", "/api/v1/users/"+createdUID, map[string]interface{}{
			"name": "Staff Updated",
		})
		_ = doRequest(appRouter, token, "DELETE", "/api/v1/users/"+createdUID, nil)
	}

	// 4. Category Handler
	catPayload := map[string]interface{}{
		"name": "Handler Cat " + uuid.New().String()[:8],
	}
	rec = doRequest(appRouter, token, "POST", "/api/v1/categories", catPayload)
	var catResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &catResp)
	catID := catResp.Data.ID

	if catID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/categories", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/categories/"+catID, nil)
		_ = doRequest(appRouter, token, "PUT", "/api/v1/categories/"+catID, map[string]interface{}{
			"description": "Updated",
		})
	}

	// 5. Product Handler
	barcode := fmt.Sprintf("BAR%09d", time.Now().UnixNano()%1000000000)
	prodPayload := map[string]interface{}{
		"name":            "Handler Prod " + uuid.New().String()[:8],
		"barcode":         barcode,
		"unit":            "pcs",
		"base_price":      15000,
		"cost_price":      10000,
		"is_stock_active": true,
		"current_stock":   50,
		"min_stock_alert": 5,
	}
	rec = doRequest(appRouter, token, "POST", "/api/v1/products", prodPayload)
	var prodResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &prodResp)
	prodID := prodResp.Data.ID

	if prodID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/products?page=1&per_page=5&search=Handler&is_active=true&is_stock_active=true", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/products/"+prodID, nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/products/search?barcode="+barcode, nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/products/search", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/products/search?barcode=NOTFOUND99999", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/products/low-stock", nil)
		_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+prodID, map[string]interface{}{
			"base_price": 16000,
		})
		_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+prodID, map[string]interface{}{
			"image_url": "",
		})
		_ = doRequest(appRouter, token, "PATCH", "/api/v1/products/"+prodID+"/toggle-active", nil)
		_ = doRequest(appRouter, token, "PATCH", "/api/v1/products/"+prodID+"/toggle-active", nil)

		// Pricing tier
		tRec := doRequest(appRouter, token, "POST", "/api/v1/products/"+prodID+"/pricing-tiers", map[string]interface{}{
			"min_quantity": 10,
			"price":        14000,
		})
		var tierResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(tRec.Body.Bytes(), &tierResp)
		if tierResp.Data.ID != "" {
			_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+prodID+"/pricing-tiers/"+tierResp.Data.ID, map[string]interface{}{
				"min_quantity": 10,
				"price":        13500,
			})
			_ = doRequest(appRouter, token, "DELETE", "/api/v1/products/"+prodID+"/pricing-tiers/"+tierResp.Data.ID, nil)
		}

		// Test delete product
		tempPRec := doRequest(appRouter, token, "POST", "/api/v1/products", map[string]interface{}{
			"name":       "Prod To Delete " + uuid.New().String()[:6],
			"unit":       "pcs",
			"base_price": 5000,
			"cost_price": 3000,
		})
		var tempPResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(tempPRec.Body.Bytes(), &tempPResp)
		if tempPResp.Data.ID != "" {
			_ = doRequest(appRouter, token, "DELETE", "/api/v1/products/"+tempPResp.Data.ID, nil)
		}

		// Multipart Create
		var multiBuf bytes.Buffer
		mw := multipart.NewWriter(&multiBuf)
		_ = mw.WriteField("data", `{"name": "Multipart Prod `+uuid.New().String()[:6]+`", "unit": "pcs", "base_price": 10000, "cost_price": 7000}`)
		_ = mw.Close()

		mReq := httptest.NewRequest("POST", "/api/v1/products", &multiBuf)
		mReq.Header.Set("Content-Type", mw.FormDataContentType())
		mReq.Header.Set("Authorization", "Bearer "+token)
		appRouter.ServeHTTP(httptest.NewRecorder(), mReq)

		// Bad multipart create (missing data)
		var badBuf bytes.Buffer
		badMw := multipart.NewWriter(&badBuf)
		_ = badMw.Close()
		bmReq := httptest.NewRequest("POST", "/api/v1/products", &badBuf)
		bmReq.Header.Set("Content-Type", badMw.FormDataContentType())
		bmReq.Header.Set("Authorization", "Bearer "+token)
		appRouter.ServeHTTP(httptest.NewRecorder(), bmReq)

		// Multipart update
		var uBuf bytes.Buffer
		uMw := multipart.NewWriter(&uBuf)
		_ = uMw.WriteField("data", `{"base_price": 17500}`)
		_ = uMw.Close()
		umReq := httptest.NewRequest("PUT", "/api/v1/products/"+prodID, &uBuf)
		umReq.Header.Set("Content-Type", uMw.FormDataContentType())
		umReq.Header.Set("Authorization", "Bearer "+token)
		appRouter.ServeHTTP(httptest.NewRecorder(), umReq)

		// Multipart update with missing data
		var uBadBuf bytes.Buffer
		uBadMw := multipart.NewWriter(&uBadBuf)
		_ = uBadMw.Close()
		ubmReq := httptest.NewRequest("PUT", "/api/v1/products/"+prodID, &uBadBuf)
		ubmReq.Header.Set("Content-Type", uBadMw.FormDataContentType())
		ubmReq.Header.Set("Authorization", "Bearer "+token)
		appRouter.ServeHTTP(httptest.NewRecorder(), ubmReq)

		// Multipart with image part (invalid extension triggers validateImage error)
		var imgMultipart bytes.Buffer
		imw := multipart.NewWriter(&imgMultipart)
		_ = imw.WriteField("data", `{"name": "Bad Ext"}`)
		part, _ := imw.CreateFormFile("image", "bad.exe")
		_, _ = part.Write([]byte("not-an-image"))
		_ = imw.Close()
		reqImg := httptest.NewRequest("POST", "/api/v1/products", &imgMultipart)
		reqImg.Header.Set("Content-Type", imw.FormDataContentType())
		reqImg.Header.Set("Authorization", "Bearer "+token)
		appRouter.ServeHTTP(httptest.NewRecorder(), reqImg)

		// Multipart update with image part (invalid extension triggers validateImage error)
		var imgUpMultipart bytes.Buffer
		uimw := multipart.NewWriter(&imgUpMultipart)
		_ = uimw.WriteField("data", `{"base_price": 1000}`)
		part2, _ := uimw.CreateFormFile("image", "bad.exe")
		_, _ = part2.Write([]byte("not-an-image"))
		_ = uimw.Close()
		reqImgUp := httptest.NewRequest("PUT", "/api/v1/products/"+prodID, &imgUpMultipart)
		reqImgUp.Header.Set("Content-Type", uimw.FormDataContentType())
		reqImgUp.Header.Set("Authorization", "Bearer "+token)
		appRouter.ServeHTTP(httptest.NewRecorder(), reqImgUp)
	}

	// 6. Cash Flow Handler
	openDrawerResp := doRequest(appRouter, token, "POST", "/api/v1/cashflow/drawer/open", map[string]interface{}{
		"opening_balance": 100000,
	})
	var dResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(openDrawerResp.Body.Bytes(), &dResp)
	openDrawerID := dResp.Data.ID

	_ = doRequest(appRouter, token, "POST", "/api/v1/cashflow", map[string]interface{}{
		"type":   "expense",
		"amount": 5000,
		"notes":  "Beli es batu",
	})
	_ = doRequest(appRouter, token, "GET", "/api/v1/cashflow/categories", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/cashflow/drawer/current", nil)

	// 7. Customer & Kasbon Handler
	phone := fmt.Sprintf("081%09d", time.Now().UnixNano()%1000000000)
	custPayload := map[string]interface{}{
		"name":         "Pelanggan Handler " + uuid.New().String()[:8],
		"phone":        phone,
		"credit_limit": 500000,
	}
	rec = doRequest(appRouter, token, "POST", "/api/v1/customers", custPayload)
	var custResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &custResp)
	custID := custResp.Data.ID

	if custID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/customers?page=1&per_page=5&search=Pelanggan&has_debt=true&is_active=true", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/customers?is_active=false", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/customers?is_active=all", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/customers/"+custID, nil)
		_ = doRequest(appRouter, token, "PUT", "/api/v1/customers/"+custID, map[string]interface{}{
			"credit_limit": 600000,
		})
		_ = doRequest(appRouter, token, "GET", "/api/v1/customers/with-debt", nil)
	}

	// 8. Transaction Handler
	if prodID != "" {
		_ = doRequest(appRouter, token, "POST", "/api/v1/transactions/calculate", map[string]interface{}{
			"items": []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
		})
		txRec := doRequest(appRouter, token, "POST", "/api/v1/transactions", map[string]interface{}{
			"items":          []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
			"payment_method": "cash",
			"amount_paid":    50000,
			"cashier_name":   "Kasir Handler",
		})
		var txResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(txRec.Body.Bytes(), &txResp)
		if txResp.Data.ID != "" {
			_ = doRequest(appRouter, token, "GET", "/api/v1/transactions?page=1&per_page=5&search=Kasir&status=completed&payment_method=cash&date_from=2026-10-01&date_to=2026-10-10", nil)
			_ = doRequest(appRouter, token, "GET", "/api/v1/transactions/"+txResp.Data.ID, nil)
			_ = doRequest(appRouter, token, "POST", "/api/v1/transactions/"+txResp.Data.ID+"/cancel", nil)
		}

		// Kasbon transaction for customer
		if custID != "" {
			_ = doRequest(appRouter, token, "POST", "/api/v1/transactions", map[string]interface{}{
				"customer_id":    custID,
				"items":          []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
				"payment_method": "kasbon",
				"amount_paid":    0,
				"cashier_name":   "Kasir Handler",
			})
			_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+custID+"?page=1&per_page=5&type=debt", nil)
			_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+custID+"?type=payment", nil)
			_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+custID+"/summary", nil)
			_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+custID+"/billing/pdf", nil)
			_ = doRequest(appRouter, token, "POST", "/api/v1/kasbon/customers/"+custID+"/payments", map[string]interface{}{
				"amount": 5000,
			})
		}
	}

	// 9. Location Handler
	locRec := doRequest(appRouter, token, "POST", "/api/v1/locations", map[string]interface{}{
		"name":         "Rak Handler " + uuid.New().String()[:4],
		"category":     "shelf",
		"width":        1.0,
		"depth":        1.0,
		"height":       2.0,
		"x_coordinate": 0.0,
		"y_coordinate": 0.0,
		"z_coordinate": 0.0,
	})
	var locResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(locRec.Body.Bytes(), &locResp)
	if locResp.Data.ID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/locations/"+locResp.Data.ID, nil)
		_ = doRequest(appRouter, token, "PUT", "/api/v1/locations/"+locResp.Data.ID, map[string]interface{}{"name": "Updated"})
		_ = doRequest(appRouter, token, "DELETE", "/api/v1/locations/"+locResp.Data.ID, nil)
	}

	// 10. Consignment Handler
	conRec := doRequest(appRouter, token, "POST", "/api/v1/consignors", map[string]interface{}{
		"name": "Titip Donat " + uuid.New().String()[:4],
	})
	var conResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(conRec.Body.Bytes(), &conResp)
	if conResp.Data.ID != "" {
		_ = doRequest(appRouter, token, "PUT", "/api/v1/consignors/"+conResp.Data.ID, map[string]interface{}{"name": "Updated Donat"})
		_ = doRequest(appRouter, token, "DELETE", "/api/v1/consignors/"+conResp.Data.ID, nil)
	}

	// 11. Refillable Handler
	if prodID != "" {
		rfRec := doRequest(appRouter, token, "POST", "/api/v1/refillables", map[string]interface{}{
			"product_id":     prodID,
			"container_type": "galon",
			"empty_count":    2,
			"full_count":     5,
		})
		var rfResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(rfRec.Body.Bytes(), &rfResp)
		if rfResp.Data.ID != "" {
			_ = doRequest(appRouter, token, "POST", "/api/v1/refillables/adjust", map[string]interface{}{
				"container_id": rfResp.Data.ID,
				"empty_change": 1,
				"full_change":  -1,
			})
			_ = doRequest(appRouter, token, "GET", "/api/v1/refillables/"+rfResp.Data.ID+"/movements", nil)
		}
	}

	// 12. POS Held Cart & Refund Handler
	if prodID != "" {
		hcRec := doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts", map[string]interface{}{
			"items":  []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
			"held_by": "Tester",
		})
		var hcResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(hcRec.Body.Bytes(), &hcResp)
		if hcResp.Data.ID != "" {
			_ = doRequest(appRouter, token, "GET", "/api/v1/pos/held-carts/"+hcResp.Data.ID, nil)
			_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts/"+hcResp.Data.ID+"/resume", nil)
		}
	}

	// 13. Stock Opname Handler
	soRec := doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions", map[string]interface{}{
		"notes": "Handler opname test",
	})
	var soResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(soRec.Body.Bytes(), &soResp)
	if soResp.Data.ID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions/"+soResp.Data.ID, nil)
		if prodID != "" {
			_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+soResp.Data.ID+"/items", map[string]interface{}{
				"product_id":     prodID,
				"physical_stock": 50,
			})
			_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions/"+soResp.Data.ID+"/variance", nil)
		}
		_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+soResp.Data.ID+"/finalize", map[string]interface{}{
			"apply_adjustments": false,
		})
	}

	// 14. Inventory Handler
	if prodID != "" {
		_ = doRequest(appRouter, token, "POST", "/api/v1/inventory/restock", map[string]interface{}{
			"product_id":    prodID,
			"quantity":      10,
			"cost_per_unit": 10000,
		})
		_ = doRequest(appRouter, token, "POST", "/api/v1/inventory/adjust", map[string]interface{}{
			"product_id": prodID,
			"quantity":   -1,
			"reason":     "Rusak",
		})
		_ = doRequest(appRouter, token, "GET", "/api/v1/inventory/"+prodID+"/movements?page=1&per_page=5", nil)
		_ = doRequest(appRouter, token, "GET", "/api/v1/inventory/restock-list/pdf", nil)
	}

	// 15. Notifications Handler
	_ = doRequest(appRouter, token, "PATCH", "/api/v1/notifications/00000000-0000-0000-0000-000000000000/read", nil)
	_ = doRequest(appRouter, token, "PATCH", "/api/v1/notifications/read-all", nil)

	// 16. Payment Handler
	_ = doRequest(appRouter, token, "POST", "/api/v1/payments/snap", map[string]interface{}{})
	_ = doRequest(appRouter, token, "POST", "/api/v1/payments/qris/charge", map[string]interface{}{})
	_ = doRequest(appRouter, token, "POST", "/api/v1/payments/00000000-0000-0000-0000-000000000000/manual-verify", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/payments/status/00000000-0000-0000-0000-000000000000", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/payments/transaction/00000000-0000-0000-0000-000000000000", nil)

	if prodID != "" {
		_ = doRequest(appRouter, token, "POST", "/api/v1/payments/snap", map[string]interface{}{
			"transaction_id": uuid.New().String(),
		})
		_ = doRequest(appRouter, token, "POST", "/api/v1/payments/qris/charge", map[string]interface{}{
			"transaction_id": uuid.New().String(),
		})
	}

	// 17. POS Refund & Discard
	if prodID != "" {
		discardCart := doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts", map[string]interface{}{
			"items":   []map[string]interface{}{{"product_id": prodID, "quantity": 1}},
			"held_by": "Tester",
		})
		var dCartResp struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		_ = json.Unmarshal(discardCart.Body.Bytes(), &dCartResp)
		if dCartResp.Data.ID != "" {
			_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts/"+dCartResp.Data.ID+"/discard", nil)
		}

		_ = doRequest(appRouter, token, "POST", "/api/v1/pos/refunds", map[string]interface{}{
			"transaction_id": uuid.New().String(),
			"refund_method":  "cash",
			"reason":         "Rusak",
			"items":          []map[string]interface{}{},
		})
	}

	// 18. Stock Opname Extra Endpoints
	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions?page=1&per_page=5&status=completed", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/shopping-list", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/near-expiry?days=45", nil)
	cancelSO := doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions", map[string]interface{}{"notes": "Cancel"})
	var cSOResp struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(cancelSO.Body.Bytes(), &cSOResp)
	if cSOResp.Data.ID != "" {
		_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+cSOResp.Data.ID+"/cancel", nil)
	}

	// 19. Customer Delete & Customers with debt
	if custID != "" {
		_ = doRequest(appRouter, token, "GET", "/api/v1/customers/with-debt", nil)
		_ = doRequest(appRouter, token, "POST", "/api/v1/kasbon/customers/"+custID+"/payments", map[string]interface{}{
			"amount": 10000,
		})
		_ = doRequest(appRouter, token, "DELETE", "/api/v1/customers/"+custID, nil)
	}

	// 20. Public Product Detail & Category Delete
	if prodID != "" {
		_ = doRequest(appRouter, "", "GET", "/public/products/"+prodID, nil)
		_ = doRequest(appRouter, "", "GET", "/public/products?page=1&per_page=5&search=Handler&sort_by=name&sort_order=desc", nil)
		_ = doRequest(appRouter, "", "GET", "/public/categories?search=Handler", nil)
	}
	if catID != "" {
		_ = doRequest(appRouter, token, "DELETE", "/api/v1/categories/"+catID, nil)
	}

	// 21. Cashflow List & Close Drawer
	_ = doRequest(appRouter, token, "GET", "/api/v1/cashflow?page=1&per_page=5&type=expense&date_from=2026-10-01T00:00:00Z", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/reports/kasbon", nil)
	if openDrawerID != "" {
		_ = doRequest(appRouter, token, "POST", "/api/v1/cashflow/drawer/close", map[string]interface{}{
			"session_id":      openDrawerID,
			"closing_balance": 100000,
		})
	}

	// 22. Realtime Events Stream Handler
	ctxEvents, cancelEvents := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancelEvents()
	reqEvents := httptest.NewRequestWithContext(ctxEvents, "GET", "/api/v1/events", nil)
	reqEvents.Header.Set("Authorization", "Bearer "+token)
	recEvents := httptest.NewRecorder()
	appRouter.ServeHTTP(recEvents, reqEvents)

	_ = ts
}

func TestDirectHandlerMethods(t *testing.T) {
	_, _, _, db := setupTestEnvironment(t)

	// Repositories
	kasbonRepo := repository.NewKasbonRepository(db)
	custRepo := repository.NewCustomerRepository(db)
	cfRepo := repository.NewCashFlowRepository(db)
	prodRepo := repository.NewProductRepository(db)
	userRepo := repository.NewUserRepository(db)
	catRepo := repository.NewCategoryRepository(db)
	locRepo := repository.NewLocationRepository(db)
	consignRepo := repository.NewConsignmentRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	posRepo := repository.NewPOSRepository(db)
	soRepo := repository.NewStockOpnameRepository(db)
	invRepo := repository.NewInventoryRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	payRepo := repository.NewPaymentRepository(db)

	cacheSvc := service.NewCacheService(nil)
	notifSvc := service.NewNotificationService(notifRepo, nil, nil)
	userSvc := service.NewUserService(userRepo)
	catSvc := service.NewCategoryService(catRepo)
	consignSvc := service.NewConsignmentService(db, consignRepo, txRepo)
	posSvc := service.NewPOSService(db, posRepo, prodRepo, txRepo, invRepo)
	soSvc := service.NewStockOpnameService(db, soRepo, prodRepo, invRepo)
	txSvc := service.NewTransactionService(db, txRepo, prodRepo, custRepo, kasbonRepo, invRepo, nil, cfRepo, notifSvc)
	paySvc := service.NewPaymentService(db, payRepo, txRepo, &config.MidtransConfig{Environment: "sandbox"})

	// Handlers
	kh := handler.NewKasbonHandler(kasbonRepo, custRepo, cfRepo)
	ch := handler.NewCustomerHandler(custRepo)
	ph := handler.NewProductHandler(prodRepo, nil, cacheSvc)
	uh := handler.NewUserHandler(userSvc)
	cath := handler.NewCategoryHandler(catSvc)
	lh := handler.NewLocationHandler(locRepo)
	conh := handler.NewConsignmentHandler(consignSvc)
	posh := handler.NewPOSHandler(posSvc)
	soh := handler.NewStockOpnameHandler(soSvc)
	txh := handler.NewTransactionHandler(txSvc, txRepo)
	payh := handler.NewPaymentHandler(paySvc)
	nh := handler.NewNotificationHandler(notifSvc)

	badReq := httptest.NewRequest("GET", "/test", bytes.NewReader([]byte("invalid json{")))
	badReq.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// 1. KasbonHandler
	kh.GetReport(rec, badReq)
	kh.GetHistory(rec, badReq)
	kh.GetSummary(rec, badReq)
	kh.RecordPayment(rec, badReq)
	kh.DownloadBillingPDF(rec, badReq)

	// 2. CustomerHandler
	ch.GetByID(rec, badReq)
	ch.Update(rec, badReq)
	ch.Delete(rec, badReq)
	ch.Create(rec, badReq)

	// 3. ProductHandler
	ph.GetByID(rec, badReq)
	ph.Update(rec, badReq)
	ph.Delete(rec, badReq)
	ph.ToggleActive(rec, badReq)
	ph.AddPricingTier(rec, badReq)
	ph.UpdatePricingTier(rec, badReq)
	ph.DeletePricingTier(rec, badReq)
	ph.Create(rec, badReq)

	// 4. UserHandler
	uh.GetByID(rec, badReq)
	uh.Update(rec, badReq)
	uh.Delete(rec, badReq)
	uh.Create(rec, badReq)

	// 5. CategoryHandler
	cath.GetByID(rec, badReq)
	cath.Update(rec, badReq)
	cath.Delete(rec, badReq)
	cath.Create(rec, badReq)

	// 6. LocationHandler
	lh.GetByID(rec, badReq)
	lh.Update(rec, badReq)
	lh.Delete(rec, badReq)
	lh.Create(rec, badReq)

	// 7. ConsignmentHandler
	conh.UpdateConsignor(rec, badReq)
	conh.DeleteConsignor(rec, badReq)
	conh.CreateConsignor(rec, badReq)

	// 8. POSHandler
	posh.GetHeldCart(rec, badReq)
	posh.ResumeCart(rec, badReq)
	posh.DiscardCart(rec, badReq)
	posh.CreateRefund(rec, badReq)
	posh.HoldCart(rec, badReq)

	// 9. StockOpnameHandler
	soh.GetSession(rec, badReq)
	soh.RecordCount(rec, badReq)
	soh.FinalizeSession(rec, badReq)
	soh.GetVarianceReport(rec, badReq)
	soh.CancelSession(rec, badReq)
	soh.StartSession(rec, badReq)

	// 10. TransactionHandler
	txh.GetByID(rec, badReq)
	txh.Cancel(rec, badReq)
	txh.Calculate(rec, badReq)
	txh.Create(rec, badReq)

	// 11. PaymentHandler
	payh.GenerateSnapToken(rec, badReq)
	payh.CreateQRISCharge(rec, badReq)
	payh.ManualVerify(rec, badReq)
	payh.GetPaymentByTransaction(rec, badReq)
	payh.CheckPaymentStatus(rec, badReq)
	payh.HandleNotification(rec, badReq)

	// 12. NotificationHandler
	nh.MarkAsRead(rec, badReq)
	nh.MarkAllAsRead(rec, badReq)
}

func TestHandlerValidationAndErrors(t *testing.T) {
	appRouter, token, _, _ := setupTestEnvironment(t)

	// Validation and error paths through HTTP
	badJSON := "invalid-json{{"
	notUUID := "not-a-uuid"
	notExistUUID := uuid.New().String()

	// Products
	_ = doRequest(appRouter, token, "POST", "/api/v1/products", map[string]interface{}{"name": ""})
	_ = doRequest(appRouter, token, "POST", "/api/v1/products", badJSON)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+notUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+notExistUUID, badJSON)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/products/"+notUUID, nil)
	_ = doRequest(appRouter, token, "PATCH", "/api/v1/products/"+notUUID+"/toggle-active", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/products/"+notUUID+"/pricing-tiers", nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+notUUID+"/pricing-tiers/"+notUUID, nil)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/products/"+notUUID+"/pricing-tiers/"+notUUID, nil)

	// Public
	_ = doRequest(appRouter, "", "GET", "/public/products/"+notUUID, nil)
	_ = doRequest(appRouter, "", "GET", "/public/products/"+notExistUUID, nil)

	// Categories
	_ = doRequest(appRouter, token, "POST", "/api/v1/categories", badJSON)
	_ = doRequest(appRouter, token, "GET", "/api/v1/categories/"+notUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/categories/"+notUUID, nil)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/categories/"+notUUID, nil)

	// Customers
	_ = doRequest(appRouter, token, "POST", "/api/v1/customers", map[string]interface{}{"name": "a"}) // too short
	_ = doRequest(appRouter, token, "POST", "/api/v1/customers", badJSON)
	_ = doRequest(appRouter, token, "GET", "/api/v1/customers/"+notUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/customers/"+notUUID, nil)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/customers/"+notUUID, nil)

	// Transactions
	_ = doRequest(appRouter, token, "POST", "/api/v1/transactions", map[string]interface{}{"items": []interface{}{}})
	_ = doRequest(appRouter, token, "POST", "/api/v1/transactions", badJSON)
	_ = doRequest(appRouter, token, "POST", "/api/v1/transactions/calculate", badJSON)
	_ = doRequest(appRouter, token, "GET", "/api/v1/transactions/"+notUUID, nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/transactions/"+notUUID+"/cancel", nil)

	// POS
	_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts", badJSON)
	_ = doRequest(appRouter, token, "GET", "/api/v1/pos/held-carts/"+notUUID, nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts/"+notUUID+"/resume", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts/"+notUUID+"/discard", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/pos/refunds", badJSON)

	// Locations
	_ = doRequest(appRouter, token, "POST", "/api/v1/locations", badJSON)
	_ = doRequest(appRouter, token, "GET", "/api/v1/locations/"+notUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/locations/"+notUUID, nil)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/locations/"+notUUID, nil)

	// Stock Opname
	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions/"+notUUID, nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+notUUID+"/items", badJSON)
	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions/"+notUUID+"/variance", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+notUUID+"/finalize", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+notUUID+"/cancel", nil)

	// Consignors
	_ = doRequest(appRouter, token, "POST", "/api/v1/consignors", badJSON)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/consignors/"+notUUID, nil)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/consignors/"+notUUID, nil)

	// Refillables
	_ = doRequest(appRouter, token, "POST", "/api/v1/refillables", badJSON)
	_ = doRequest(appRouter, token, "POST", "/api/v1/refillables/adjust", badJSON)

	// Users
	_ = doRequest(appRouter, token, "POST", "/api/v1/users", map[string]interface{}{"name": "u"})
	_ = doRequest(appRouter, token, "GET", "/api/v1/users/"+notUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/users/"+notUUID, nil)
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/users/"+notUUID, nil)

	// Kasbon
	_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+notUUID, nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+notUUID+"/summary", nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/kasbon/customers/"+notUUID+"/billing/pdf", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/kasbon/customers/"+notUUID+"/payments", badJSON)

	// Auth Refresh
	_ = doRequest(appRouter, "", "POST", "/auth/refresh", map[string]string{"refresh_token": ""})
	_ = doRequest(appRouter, "", "POST", "/auth/refresh", map[string]string{"refresh_token": "bad"})

	// Not Found Paths (404)
	_ = doRequest(appRouter, token, "GET", "/api/v1/categories/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/categories/"+notExistUUID, map[string]string{"name": "n"})
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/categories/"+notExistUUID, nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/customers/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/customers/"+notExistUUID, map[string]interface{}{"credit_limit": 100})
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/customers/"+notExistUUID, nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/locations/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/locations/"+notExistUUID, map[string]string{"name": "n"})
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/locations/"+notExistUUID, nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/products/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/products/"+notExistUUID, map[string]interface{}{"base_price": 100})
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/products/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "PATCH", "/api/v1/products/"+notExistUUID+"/toggle-active", nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/transactions/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/transactions/"+notExistUUID+"/cancel", nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/pos/held-carts/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts/"+notExistUUID+"/resume", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/pos/held-carts/"+notExistUUID+"/discard", nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "GET", "/api/v1/stock-opname/sessions/"+notExistUUID+"/variance", nil)
	_ = doRequest(appRouter, token, "POST", "/api/v1/stock-opname/sessions/"+notExistUUID+"/cancel", nil)

	_ = doRequest(appRouter, token, "PUT", "/api/v1/consignors/"+notExistUUID, map[string]string{"name": "n"})
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/consignors/"+notExistUUID, nil)

	_ = doRequest(appRouter, token, "GET", "/api/v1/users/"+notExistUUID, nil)
	_ = doRequest(appRouter, token, "PUT", "/api/v1/users/"+notExistUUID, map[string]string{"name": "n"})
	_ = doRequest(appRouter, token, "DELETE", "/api/v1/users/"+notExistUUID, nil)
}
