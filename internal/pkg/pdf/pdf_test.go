package pdf_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/eveeze/warung-backend/internal/pkg/pdf"
)

func TestGenerateBillingPDF(t *testing.T) {
	now := time.Now()
	data := pdf.BillingData{
		CustomerName:   "Budi Santoso",
		InvoiceNumber:  "INV-TEST-001",
		Date:           now,
		PeriodStart:    &now,
		PeriodEnd:      &now,
		OpeningBalance: 50000,
		Transactions: []pdf.BillingTransaction{
			{Date: now, Description: "Beras 5kg", Type: "debt", Amount: 75000, Balance: 125000},
			{Date: now, Description: "Cicilan 1", Type: "payment", Amount: 25000, Balance: 100000},
		},
		EndingBalance: 100000,
		StoreName:     "Warung Kelontong Berkah",
		StoreAddress:  "Jl. Merdeka No. 45",
		PaymentInst:   "BCA 123456789 a.n. Warung",
	}

	doc, err := pdf.GenerateBillingPDF(data)
	if err != nil {
		t.Fatalf("failed to generate billing pdf: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("failed to output pdf buffer: %v", err)
	}
	if buf.Len() < 500 {
		t.Errorf("expected generated PDF bytes > 500, got %d", buf.Len())
	}
}

func TestGenerateRestockPDF(t *testing.T) {
	now := time.Now()
	data := pdf.RestockData{
		StoreName:    "Warung Kelontong Berkah",
		StoreAddress: "Jl. Merdeka No. 45",
		GeneratedAt:  now,
		Items: []pdf.RestockItem{
			{
				ProductName:  "Minyak Goreng 1L",
				CurrentStock: 2,
				MinStock:     10,
				Deficit:      8,
				Unit:         "pch",
				CostPrice:    14000,
			},
			{
				ProductName:  "Gula Pasir 1kg",
				CurrentStock: 0,
				MinStock:     20,
				Deficit:      20,
				Unit:         "kg",
				CostPrice:    15000,
			},
		},
	}

	doc, err := pdf.GenerateRestockPDF(data)
	if err != nil {
		t.Fatalf("failed to generate restock pdf: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		t.Fatalf("failed to output pdf buffer: %v", err)
	}
	if buf.Len() < 500 {
		t.Errorf("expected generated PDF bytes > 500, got %d", buf.Len())
	}
}
