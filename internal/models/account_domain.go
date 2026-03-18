package models

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type AccountFlag = int64

const (
	AccountFlagIsDefault AccountFlag = 1 << 0
)

type AccountType string

const (
	AccountTypeBank       AccountType = "bank"
	AccountTypeCash       AccountType = "cash"
	AccountTypeEWallet    AccountType = "ewallet"
	AccountTypeCreditCard AccountType = "credit_card"
	AccountTypeInvestment AccountType = "investment"
	AccountTypeLoan       AccountType = "loan"
	AccountTypeOther      AccountType = "other"
)

type Account struct {
	Id                string          `json:"id" gorm:"column:id"`
	UserId            string          `json:"user_id" gorm:"column:user_id;foreignKey:UserId;references:Id"`
	Name              string          `json:"name" gorm:"column:name"`
	Currency          string          `json:"currency" gorm:"column:currency"` // Currency Code
	Balance           decimal.Decimal `json:"balance" gorm:"column:balance"`
	CreatedAt         time.Time       `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time       `json:"updated_at" gorm:"column:updated_at"`
	DeletedAt         sql.NullTime    `json:"deleted_at" gorm:"column:deleted_at"`
	Type              AccountType     `json:"type" gorm:"column:type"`
	Note              string          `json:"note" gorm:"column:note"`
	Flag              AccountFlag     `json:"flag" gorm:"column:flag"` // Flag for default account
	AccountNumber     string          `json:"account_number" gorm:"column:account_number"`
	Iban              string          `json:"iban" gorm:"column:iban"`
	DisplayOrder      sql.NullInt32   `json:"display_order" gorm:"column:display_order"`
	FirsTransactionAt sql.NullTime    `json:"first_transaction_at" gorm:"column:first_transaction_at"`
	LastTransactionAt sql.NullTime    `json:"last_transaction_at" gorm:"column:last_transaction_at"`
}

func (a *Account) TableName() string {
	return "accounts"
}

func (a *Account) IsDefault() bool {
	return a.Flag&AccountFlagIsDefault == AccountFlagIsDefault
}
