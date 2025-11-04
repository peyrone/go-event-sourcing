package util

import (
	"fmt"
	"time"
)

// NewID generates a simple unique ID using a timestamp
func NewID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
