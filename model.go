package api

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type Product struct {
	SKU   string `json:"sku"`
	Attrs Attrs  `json:"attrs"`
}

// Attrs is a map of key-value pairs associated with a product.
type Attrs map[string]string

// Value implements the SQL driver.Value interface
func (a *Attrs) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface
func (a *Attrs) Scan(v interface{}) error {
	b, ok := v.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
