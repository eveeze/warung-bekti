package pdf

import (
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

type BillingData struct {
	CustomerName    string
	InvoiceNumber   string
	Date            time.Time
	PeriodStart     *time.Time
	PeriodEnd       *time.Time
	OpeningBalance  int64
	Transactions    []BillingTransaction
	EndingBalance   int64
	StoreName       string
	StoreAddress    string
	PaymentInst     string
}

type BillingTransaction struct {
	Date        time.Time
	Description string
	Type        string // "debt" or "payment"
	Amount      int64
	Balance     int64
}

// GenerateBillingPDF creates a neat PDF billing statement
func GenerateBillingPDF(data BillingData) (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// -- Colors and Fonts --
	darkGray := []int{50, 50, 50}
	lightGray := []int{240, 240, 240}
	primaryColor := []int{0, 0, 0} // Black for professional look, or maybe a nice blue

	// -- Header --
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Cell(0, 10, data.StoreName)
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
	pdf.Cell(0, 5, data.StoreAddress)
	pdf.Ln(15)

	// -- Title & Info --
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 10, "TAGIHAN KASBON (BILLING STATEMENT)")
	pdf.Ln(12)

	// Info Grid
	pdf.SetFont("Arial", "", 10)
	
	// Left Column: Customer Info
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, "Kepada:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, data.CustomerName)
	pdf.Ln(5)
	
	// Right Column: Invoice Info (Manually positioned or just sequential)
	// Let's keep it sequential for simplicity but use Cells basically
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, "Nomor:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, data.InvoiceNumber)
	pdf.Ln(5)
	
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(30, 5, "Tanggal:")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, data.Date.Format("02 Jan 2006"))
	pdf.Ln(10)

	// -- Table --
	
	// Table Config
	colWidths := []float64{30, 80, 25, 25, 30} // Total 190 (A4 is 210, margins 15+15=30, usable 180. Adjust)
	// Let's use 180 width: 30 (Date) + 70 (Desc) + 25 (Debit) + 25 (Credit) + 30 (Balance)
	colWidths = []float64{30, 70, 25, 25, 30}
	headers := []string{"Tanggal", "Keterangan", "Tambah", "Bayar", "Saldo"}

	// Header Row
	pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// Opening Balance Row
	pdf.SetFont("Arial", "I", 10)
	pdf.CellFormat(colWidths[0], 8, "", "L", 0, "", false, 0, "")
	pdf.CellFormat(colWidths[1], 8, "Saldo Awal", "", 0, "L", false, 0, "")
	pdf.CellFormat(colWidths[2], 8, "-", "", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[3], 8, "-", "", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[4], 8, formatMoney(data.OpeningBalance), "R", 0, "R", false, 0, "")
	pdf.Ln(-1)
	
	// Data Rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(50, 50, 50)
	
	for _, tx := range data.Transactions {
		pdf.CellFormat(colWidths[0], 8, tx.Date.Format("02/01/06"), "LR", 0, "L", false, 0, "")
		
		desc := truncateString(tx.Description, 35) // Simple truncate
		pdf.CellFormat(colWidths[1], 8, desc, "R", 0, "L", false, 0, "")
		
		debtStr := "-"
		payStr := "-"
		
		if tx.Type == "debt" {
			debtStr = formatMoney(tx.Amount)
		} else {
			payStr = formatMoney(tx.Amount)
		}
		
		pdf.CellFormat(colWidths[2], 8, debtStr, "R", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[3], 8, payStr, "R", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[4], 8, formatMoney(tx.Balance), "R", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	// Bottom line of table
	pdf.CellFormat(180, 0, "", "T", 1, "", false, 0, "")

	// -- Summary / Total --
	pdf.Ln(5)
	
	// Right aligned summary
	pdf.SetX(130) // Move to right side
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(30, 10, "Total Tagihan:", "", 0, "R", false, 0, "")
	pdf.CellFormat(35, 10, formatMoney(data.EndingBalance), "", 1, "R", false, 0, "")

	// -- Footer / Payment Instructions --
	pdf.SetY(240) // Bottom of page usually
	
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(0, 5, "Instruksi Pembayaran:")
	pdf.Ln(6)
	
	pdf.SetFont("Arial", "", 9)
	pdf.MultiCell(0, 5, data.PaymentInst, "", "L", false)
	
	pdf.Ln(10)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 5, fmt.Sprintf("Dicetak pada: %s", time.Now().Format("02 Jan 2006 15:04:05")))

	return pdf, nil
}

func formatMoney(amount int64) string {
	// Simple formatter (Rp 1.000.000)
	// In production, use a proper library or robust helper
	s := fmt.Sprintf("%d", amount)
	if amount < 0 {
		s = fmt.Sprintf("%d", -amount)
	}
	
	// Add dots
	n := len(s)
	if n <= 3 {
		s = "Rp " + s
	} else {
		var result []byte
		for i, c := range s {
			if (n-i)%3 == 0 && i != 0 {
				result = append(result, '.')
			}
			result = append(result, byte(c))
		}
		s = "Rp " + string(result)
	}
	
	if amount < 0 {
		return "-" + s
	}
	return s
}

func truncateString(s string, max int) string {
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
}
