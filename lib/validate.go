package lib

import (
	"fmt"
	"strconv"
)

func validateSystemID(id string) error {
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return fmt.Errorf("invalid system ID %q: must be numeric", id)
	}
	return nil
}
