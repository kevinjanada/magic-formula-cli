package models

import "time"

type Stock struct {
	BaseModel
	Code string
	Name string
	ListingDate time.Time
	Shares int64
	ListingBoard string
}