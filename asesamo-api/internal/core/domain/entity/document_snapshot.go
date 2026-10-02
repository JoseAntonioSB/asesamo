package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// DocumentSnapshotItem represents the compliance state of a single document
// at the moment an access event occurred.
type DocumentSnapshotItem struct {
	DocumentTypeCode string  `json:"document_type_code"`
	DocumentName     string  `json:"document_name"`
	AppliesTo        string  `json:"applies_to"`
	Status           string  `json:"status"`
	IsRequired       bool    `json:"is_required"`
	IsValid          bool    `json:"is_valid"`
	ExpiresAt        *string `json:"expires_at"`
	FailureReason    *string `json:"failure_reason"`
}

// DocumentSnapshot is a slice of DocumentSnapshotItem supporting SQL driver/scanner
// for direct persistence into PostgreSQL JSONB columns.
type DocumentSnapshot []DocumentSnapshotItem

// Value implements the database/sql/driver.Valuer interface.
func (d DocumentSnapshot) Value() (driver.Value, error) {
	if d == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("entity: failed to marshal DocumentSnapshot: %w", err)
	}
	return string(bytes), nil
}

// Scan implements the database/sql.Scanner interface.
func (d *DocumentSnapshot) Scan(src interface{}) error {
	if src == nil {
		*d = DocumentSnapshot{}
		return nil
	}

	var sourceBytes []byte
	switch s := src.(type) {
	case []byte:
		sourceBytes = s
	case string:
		sourceBytes = []byte(s)
	default:
		return fmt.Errorf("entity: unsupported type for DocumentSnapshot: %T", src)
	}

	if len(sourceBytes) == 0 {
		*d = DocumentSnapshot{}
		return nil
	}

	var items []DocumentSnapshotItem
	if err := json.Unmarshal(sourceBytes, &items); err != nil {
		return fmt.Errorf("entity: failed to unmarshal DocumentSnapshot: %w", err)
	}

	*d = items
	return nil
}
