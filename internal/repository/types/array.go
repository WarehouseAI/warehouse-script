package types

import (
	"database/sql/driver"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
)

type Int64Array []int64

func (a *Int64Array) Scan(src interface{}) error {
	return pgtype.NewMap().SQLScanner((*[]int64)(a)).Scan(src)
}

func (a *Int64Array) Value() (driver.Value, error) {
	t, ok := pgtype.NewMap().TypeForValue(*a)
	if !ok {
		return nil, errors.New("failed to find type for value int64 array")
	}
	buf := make([]byte, 0, 128)
	var err error
	buf, err = pgtype.NewMap().Encode(t.OID, pgtype.TextFormatCode, *a, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

type Int32Array []int32

func (a *Int32Array) Scan(src interface{}) error {
	return pgtype.NewMap().SQLScanner((*[]int32)(a)).Scan(src)
}

type StringArray []string

func (a *StringArray) Scan(src interface{}) error {
	return pgtype.NewMap().SQLScanner((*[]string)(a)).Scan(src)
}

func (a *StringArray) Value() (driver.Value, error) {
	t, ok := pgtype.NewMap().TypeForValue(*a)
	if !ok {
		return nil, errors.New("failed to find type for value string array")
	}
	buf := make([]byte, 0, 128)
	var err error
	buf, err = pgtype.NewMap().Encode(t.OID, pgtype.TextFormatCode, *a, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

type Float64Array []float64

func (a *Float64Array) Scan(src interface{}) error {
	return pgtype.NewMap().SQLScanner((*[]float64)(a)).Scan(src)
}

func (a *Float64Array) Value() (driver.Value, error) {
	t, ok := pgtype.NewMap().TypeForValue(*a)
	if !ok {
		return nil, errors.New("failed to find type for value float64 array")
	}
	buf := make([]byte, 0, 128)
	var err error
	buf, err = pgtype.NewMap().Encode(t.OID, pgtype.TextFormatCode, *a, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

type BoolArray []bool

func (a *BoolArray) Scan(src interface{}) error {
	return pgtype.NewMap().SQLScanner((*[]bool)(a)).Scan(src)
}

func (a *BoolArray) Value() (driver.Value, error) {
	t, ok := pgtype.NewMap().TypeForValue(*a)
	if !ok {
		return nil, errors.New("failed to find type for value bool array")
	}
	buf := make([]byte, 0, 128)
	var err error
	buf, err = pgtype.NewMap().Encode(t.OID, pgtype.TextFormatCode, *a, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
