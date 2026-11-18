package shared_test

import (
	"testing"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIDRequiresNonBlankValue(t *testing.T) {
	for _, bad := range []string{"", "   ", "\t"} {
		_, err := shared.NewUserID(bad)
		assert.ErrorIs(t, err, shared.ErrInvalidInput, "%q", bad)
	}
	u, err := shared.NewUserID("user-1")
	require.NoError(t, err)
	assert.Equal(t, "user-1", u.String())
	assert.False(t, u.IsZero())
}

func TestZeroUserIDIsZero(t *testing.T) {
	assert.True(t, shared.UserID{}.IsZero())
}
