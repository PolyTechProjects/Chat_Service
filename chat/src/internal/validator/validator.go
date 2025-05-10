package validator

import (
	"fmt"
	"strings"
)

func ValidateChatName(name string) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("name %s is blank", name)
	}
	if len(name) < 1 || len(name) > 52 {
		return fmt.Errorf("name %s is too short or too long", name)
	}
	for _, c := range name {
		if c <= 0x1F {
			return fmt.Errorf("name %s contains forbidden characters", name)
		}
	}
	return nil
}
