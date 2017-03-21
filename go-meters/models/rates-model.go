package models

import "time"

type Rate struct {
	Currency string    `json:"currency"`
	Date     time.Time `json:"date"`
	Value    float64   `json:"value"`
}
