package product

import (
	"strings"

	"github.com/google/uuid"
)

func genID() string {
	return strings.ToUpper(uuid.NewString()[:8])
}
