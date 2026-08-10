package ports

import "time"

// Clock abstracts time for testing
type Clock interface {
	Now() time.Time
}
