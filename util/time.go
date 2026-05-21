package util

import (
	"fmt"
	"time"
)

func DisplayTime(t time.Time) string {
	return fmt.Sprintf(t.Format("15:04:05"))
}
