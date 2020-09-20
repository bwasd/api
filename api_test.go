package api

import (
	"testing"
)

var products = []Product{
	Product{
		SKU:   "foo",
		Attrs: map[string]string{"a": "b", "c": "d"},
	},
	Product{
		SKU:   "bar",
		Attrs: map[string]string{"a": "b", "c": "d"},
	},
	Product{
		SKU:   "baz",
		Attrs: map[string]string{"a": "b", "c": "d"},
	},
	Product{
		SKU:   "foobar",
		Attrs: map[string]string{"a": "b", "c": "d"},
	},
}

func queryProductBySKU(SKU string) *Product {
	if SKU == "" {
		return nil
	}
	for _, p := range products {
		if p.SKU == SKU {
			return &p
		}
	}
	return nil
}

func createProduct(SKU string) error {
	if SKU == "" {
		return ErrSKUName
	}
	if p := queryProductBySKU(SKU); p != nil {
		return ErrSKUConflict
	}
	p := Product{SKU: SKU}
	products = append(products, p)
	return nil
}

func TestCreateProduct(t *testing.T) {
	var s = "kek"
	if err := createProduct("kek"); err != nil {
		t.Fatalf("creating product %s should not fail", s)
	}
}

func TestQueryProduct(t *testing.T) {
	var s = "bar"
	if v := queryProductBySKU(s); v == nil {
		t.Fatalf("expected product %s; got nil", s)
	}
}
