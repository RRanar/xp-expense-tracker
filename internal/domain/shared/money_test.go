package shared_test

import (
	"regexp"
	"testing"
	"testing/quick"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrencyAcceptsOnlyThreeUppercaseLetters(t *testing.T) {
	for _, ok := range []string{"USD", "EUR", "UAH"} {
		c, err := shared.NewCurrency(ok)
		require.NoError(t, err, ok)
		assert.Equal(t, ok, c.String())
	}
	for _, bad := range []string{"", "US", "USDD", "usd", "U5D", " USD", "US D"} {
		_, err := shared.NewCurrency(bad)
		assert.ErrorIs(t, err, shared.ErrInvalidInput, bad)
	}
}

var threeUpper = regexp.MustCompile(`^[A-Z]{3}$`)

func TestCurrencyValidityMatchesFormatForAnyString(t *testing.T) {
	prop := func(s string) bool {
		c, err := shared.NewCurrency(s)
		if threeUpper.MatchString(s) {
			return err == nil && c.String() == s
		}
		return err != nil
	}
	assert.NoError(t, quick.Check(prop, nil))
}

func TestMoneyKeepsMinorUnitsAndCurrencyForAnyAmount(t *testing.T) {
	usd, err := shared.NewCurrency("USD")
	require.NoError(t, err)
	prop := func(minor int64) bool {
		m := shared.NewMoney(minor, usd)
		return m.Minor() == minor && m.Currency() == usd
	}
	assert.NoError(t, quick.Check(prop, nil))
}

func TestMoneyIsComparedByValue(t *testing.T) {
	usd, _ := shared.NewCurrency("USD")
	eur, _ := shared.NewCurrency("EUR")
	assert.Equal(t, shared.NewMoney(100, usd), shared.NewMoney(100, usd))
	assert.NotEqual(t, shared.NewMoney(100, usd), shared.NewMoney(100, eur))
}
