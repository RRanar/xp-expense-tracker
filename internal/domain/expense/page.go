package expense

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/RRanar/xp-expense-tracker/internal/domain/shared"
)

const (
	DefaultPageSize = 50
	MaxPageSize     = 200
)

type Filter struct {
	Category string
	From     shared.Date
	To       shared.Date
}

type Page struct {
	Size  int
	Token string
}

func (p Page) EffectiveSize() int {
	switch {
	case p.Size <= 0:
		return DefaultPageSize
	case p.Size > MaxPageSize:
		return MaxPageSize
	default:
		return p.Size
	}
}

// Cursor is the position after the last item of a page in the (date, id) ordering.
type Cursor struct {
	Date shared.Date
	ID   shared.ID
}

func EncodeCursor(c Cursor) string {
	return base64.RawURLEncoding.EncodeToString([]byte(c.Date.String() + "|" + c.ID.String()))
}

func DecodeCursor(token string) (Cursor, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, invalidToken()
	}
	datePart, idPart, found := strings.Cut(string(raw), "|")
	if !found || idPart == "" {
		return Cursor{}, invalidToken()
	}
	date, err := shared.DateFromString(datePart)
	if err != nil {
		return Cursor{}, invalidToken()
	}
	return Cursor{Date: date, ID: shared.IDFromString(idPart)}, nil
}

func invalidToken() error {
	return fmt.Errorf("%w: page token is not valid", shared.ErrInvalidInput)
}
