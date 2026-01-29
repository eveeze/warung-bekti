package pdf

import (
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
)

// RestockItem represents a product that needs restocking
type RestockItem struct {
	ProductName  string
	CurrentStock int
	MinStock     int
	Deficit      int
	Unit         string
	CostPrice    int64
}

// RestockData holds data for generating restock PDF
type RestockData struct {
	StoreName    string
	StoreAddress string
	GeneratedAt  time.Time
	Items        []RestockItem
}

// GenerateRestockPDF creates a well-formatted PDF for restock list
func GenerateRestockPDF(data RestockData) (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// Colors
	darkGray := []int{50, 50, 50}
	lightGray := []int{240, 240, 240}

	// Header
	pdf.SetFont("Arial", "B", 20)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 10, data.StoreName)
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(darkGray[0], darkGray[1], darkGray[2])
	pdf.Cell(0, 5, data.StoreAddress)
	pdf.Ln(15)

	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(0, 0, 0)
	pdf.Cell(0, 10, "DAFTAR RESTOCK BARANG (LOW STOCK)")
	pdf.Ln(12)

	// Generated date
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 5, fmt.Sprintf("Dicetak: %s", data.GeneratedAt.Format("02 Jan 2006 15:04")))
	pdf.Ln(10)

	// Table columns
	// Widths: Name(60), Current(25), Min(25), Deficit(25), Order(30), Unit(15) = 180mm
	colWidths := []float64{60, 25, 25, 25, 30, 15}
	headers := []string{"Nama Produk", "Stok", "Min", "Kurang", "Order", "Satuan"}

	// Header row
	pdf.SetFillColor(lightGray[0], lightGray[1], lightGray[2])
	pdf.SetFont("Arial", "B", 10)
	pdf.SetTextColor(0, 0, 0)

	for i, h := range headers {
		align := "C"
		if i == 0 {
			align = "L"
		}
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, align, true, 0, "")
	}
	pdf.Ln(-1)

	// Data rows
	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(50, 50, 50)

	totalItems := 0
	totalDeficit := 0

	for _, item := range data.Items {
		// Product name (truncate if too long)
		name := truncateString(item.ProductName, 30)
		pdf.CellFormat(colWidths[0], 8, name, "LR", 0, "L", false, 0, "")

		// Current stock
		pdf.CellFormat(colWidths[1], 8, fmt.Sprintf("%d", item.CurrentStock), "R", 0, "C", false, 0, "")

		// Min stock
		pdf.CellFormat(colWidths[2], 8, fmt.Sprintf("%d", item.MinStock), "R", 0, "C", false, 0, "")

		// Deficit (highlight negatives)
		deficitStr := fmt.Sprintf("%d", item.Deficit)
		pdf.CellFormat(colWidths[3], 8, deficitStr, "R", 0, "C", false, 0, "")

		// Suggested order (rounded up to nearest 5 or 10)
		suggestedOrder := calculateSuggestedOrder(item.Deficit, item.MinStock)
		pdf.CellFormat(colWidths[4], 8, fmt.Sprintf("%d", suggestedOrder), "R", 0, "C", false, 0, "")

		// Unit
		pdf.CellFormat(colWidths[5], 8, item.Unit, "R", 0, "C", false, 0, "")

		pdf.Ln(-1)

		totalItems++
		totalDeficit += item.Deficit
	}

	// Bottom border
	pdf.CellFormat(180, 0, "", "T", 1, "", false, 0, "")

	// Summary
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 11)
	pdf.Cell(0, 8, fmt.Sprintf("Total Produk yang Perlu Restock: %d item", totalItems))
	pdf.Ln(6)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Total Kekurangan Stok: %d unit", totalDeficit))

	// Footer note
	pdf.SetY(260)
	pdf.SetFont("Arial", "I", 9)
	pdf.MultiCell(0, 5, "Catatan: Kolom 'Order' adalah saran jumlah pembelian (sudah dibulatkan untuk kemudahan). Sesuaikan dengan kebutuhan dan modal yang tersedia.", "", "L", false)

	return pdf, nil
}

// calculateSuggestedOrder rounds up deficit to a convenient number
func calculateSuggestedOrder(deficit, minStock int) int {
	// Add safety margin (20% of min stock or minimum 2)
	safetyMargin := minStock / 5
	if safetyMargin < 2 {
		safetyMargin = 2
	}

	suggested := deficit + safetyMargin

	// Round up to nearest 5 for small quantities, 10 for larger
	if suggested <= 20 {
		return ((suggested + 4) / 5) * 5
	}
	return ((suggested + 9) / 10) * 10
}
