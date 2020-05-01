package models

import "time"

type StatementRowFact struct {
	StatementRow StatementRow `gorm:"foreignkey:statement_row_id"`
	Stock        Stock        `gorm:"foreignkey:stock_id"`
	Date         time.Time
	Amount       Decimal
}
