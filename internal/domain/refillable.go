package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefillableContainer struct {
	ID            uuid.UUID `json:"id"`
	ProductID     uuid.UUID `json:"product_id"`
	ContainerType string    `json:"container_type"`
	EmptyCount    int       `json:"empty_count"`
	FullCount     int       `json:"full_count"`
	Notes         *string   `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	Product *Product `json:"product,omitempty"`
}

type ContainerMovementType string

const (
	ContainerMovementSaleExchange    ContainerMovementType = "sale_exchange"
	ContainerMovementRestockExchange ContainerMovementType = "restock_exchange"
	ContainerMovementPurchaseEmpty   ContainerMovementType = "purchase_empty"
	ContainerMovementPurchaseFull    ContainerMovementType = "purchase_full"
	ContainerMovementReturnEmpty     ContainerMovementType = "return_empty"
	ContainerMovementAdjustment      ContainerMovementType = "adjustment"
)

type ContainerMovement struct {
	ID            uuid.UUID             `json:"id"`
	ContainerID   uuid.UUID             `json:"container_id"`
	Type          ContainerMovementType `json:"type"`
	EmptyChange   int                   `json:"empty_change"`
	FullChange    int                   `json:"full_change"`
	EmptyBefore   int                   `json:"empty_before"`
	EmptyAfter    int                   `json:"empty_after"`
	FullBefore    int                   `json:"full_before"`
	FullAfter     int                   `json:"full_after"`
	ReferenceType *string               `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID            `json:"reference_id,omitempty"`
	Notes         *string               `json:"notes,omitempty"`
	CreatedBy     *string               `json:"created_by,omitempty"`
	CreatedAt     time.Time             `json:"created_at"`
}
