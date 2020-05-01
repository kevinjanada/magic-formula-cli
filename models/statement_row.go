package models

type StatementRow struct {
	Statement         Statement         `gorm:"foreignkey:statement_id"`
	StatementRowTitle StatementRowTitle `gorm:"foreignkey:statement_row_title_id"`
	Order             int
	Description       string
	Properties        string
}
