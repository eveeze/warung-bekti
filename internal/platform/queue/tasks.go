package queue

// Task Types
const (
	TypeNotificationSend = "notification:send"
	TypeLowStockAlert    = "notification:low_stock"
	TypeNewTransaction   = "notification:new_transaction"
)

// Task Payloads

// PayloadNotificationSend is the generic payload for reducing specific tasks into a send action
type PayloadNotificationSend struct {
	UserID  string            `json:"user_id,omitempty"`
	Title   string            `json:"title"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}

type PayloadLowStock struct {
	ProductID   string `json:"product_id"`
	ProductName string `json:"product_name"`
	CurrentStock int    `json:"current_stock"`
	MinStock     int    `json:"min_stock"`
}

type PayloadNewTransaction struct {
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	CashierName   string `json:"cashier_name"`
}
