package shared

import "fmt"

// Currency is a validated three-letter uppercase code (for example USD).
type Currency struct {
	code string
}

func NewCurrency(code string) (Currency, error) {
	if len(code) != 3 {
		return Currency{}, fmt.Errorf("%w: currency %q must be three uppercase letters", ErrInvalidInput, code)
	}
	for i := 0; i < len(code); i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return Currency{}, fmt.Errorf("%w: currency %q must be three uppercase letters", ErrInvalidInput, code)
		}
	}
	return Currency{code: code}, nil
}

func (c Currency) String() string { return c.code }
func (c Currency) IsZero() bool   { return c.code == "" }

// Money is an amount in minor units (for example cents) plus a currency.
// Sign rules belong to the aggregate that uses it; arithmetic with overflow
// checks is added when it is first needed.
type Money struct {
	minor    int64
	currency Currency
}

func NewMoney(minor int64, currency Currency) Money {
	return Money{minor: minor, currency: currency}
}

func (m Money) Minor() int64       { return m.minor }
func (m Money) Currency() Currency { return m.currency }
