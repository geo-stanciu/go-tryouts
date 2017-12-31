package models

import "time"

// Rate - exchange rate struct
type Rate struct {
	ReferenceCurrency string    `json:"reference_currency"`
	Currency          string    `json:"currency"`
	Date              time.Time `json:"date"`
	Value             float64   `json:"value"`
}
