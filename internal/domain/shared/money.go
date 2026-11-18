package shared

import "fmt"

type Currency struct {
	code string
}

func NewCurrency(code string) (Currency, error) {
	if len(code) != 3 {
		return Currency{}, invalidCurrency(code)
	}
	for i := 0; i < len(code); i++ {
		if code[i] < 'A' || code[i] > 'Z' {
			return Currency{}, invalidCurrency(code)
		}
	}
	return Currency{code: code}, nil
}

func invalidCurrency(code string) error {
	return fmt.Errorf("%w: currency %q must be three uppercase letters", ErrInvalidInput, code)
}

func (c Currency) String() string { return c.code }
func (c Currency) IsZero() bool   { return c.code == "" }

// Money does not enforce a sign: that rule belongs to the aggregate holding it.
type Money struct {
	minor    int64
	currency Currency
}

func NewMoney(minor int64, currency Currency) Money {
	return Money{minor: minor, currency: currency}
}

func (m Money) Minor() int64       { return m.minor }
func (m Money) Currency() Currency { return m.currency }
