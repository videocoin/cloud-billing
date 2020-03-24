package v1

import (
	"database/sql/driver"
	"errors"
)

func (t TransactionType) Value() (driver.Value, error) {
	return TransactionType_name[int32(t)], nil
}

func (t *TransactionType) Scan(src interface{}) error {
	tid, ok := src.([]byte)
	if !ok {
		return errors.New("type assertion .([]byte) failed")
	}

	*t = TransactionType(TransactionType_value[string(tid)])

	return nil
}
