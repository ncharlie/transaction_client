package transaction_client

import (
	"github.com/go-playground/validator/v10"
)

type TransactionStatus string

// The status of Transaction is one of these status string.
const (
	StatusConfirm TransactionStatus = "CONFIRMED"
	StatusFailed  TransactionStatus = "FAILED"
	StatusPending TransactionStatus = "PENDING"
	StatusDNE     TransactionStatus = "DNE"
	statusInit    TransactionStatus = "INIT"
)

var validate = validator.New()

type transaction struct {
	Symbol    string `json:"symbol" validate:"required"`
	Price     uint64 `json:"price" validate:"required"`
	Timestamp uint64 `json:"timestamp" validate:"required"`

	hash   string            `json:"-"`
	status TransactionStatus `json:"-"`
}

// NewTransaction validates the input values,
// returns a new instance of `transaction` to be broadcast or polling.
func NewTransaction(symbol string, price, timestamp uint64) (*transaction, error) {
	t := &transaction{
		Symbol:    symbol,
		Price:     price,
		Timestamp: timestamp,
		status:    statusInit,
	}
	if err := validate.Struct(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *transaction) Hash() string {
	return t.hash
}

func (t *transaction) Status() TransactionStatus {
	return t.status
}
