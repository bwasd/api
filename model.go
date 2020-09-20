package api

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Product represents a product from an inventory. A product is distinguished by
// a unique string identifier, an SKU (stock keeping unit).
type Product struct {
	SKU   string `json:"sku"`
	Attrs Attrs  `json:"attrs"`
}

// Attrs is a map of key-value pairs associated with a product.
//
// An attribute key must be valid UTF-8 and not contain any non-ASCII unicode
// escapes or U+0000.  Attribute key-values are always automatically coerced to
// strings.  If values exist for attributes of a given product, the entries for
// those attributes are updated, replacing all attribute values from previous
// calls.
type Attrs map[string]string

// LookupAttr retrieves the value of the attribute named by key. If the key is
// present in the Attribute, the value (which may be empty) is returned and the
// boolean is true. Otherwise the returned value will be empty and the boolean
// will be false.
func (p *Product) LookupAttr(key string) (string, bool) {
	return p.getAttr(key)
}

// GetAttr retrieves the value of the attribute named by key. If the attribute
// is missing, an empty string is returned. To distinguish an empty value and a
// non-existing (undefined) attribute, use LookupAttr instead of calling
// GetAttr.
func (p *Product) GetAttr(key string) string {
	v, _ := p.getAttr(key)
	return v
}

func (p *Product) getAttr(key string) (string, bool) {
	if val, ok := p.Attrs[key]; ok {
		return val, true
	}
	return "", false
}

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
