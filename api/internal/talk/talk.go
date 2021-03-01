package talk

import "github.com/google/uuid"

type Talk struct {
	ID   uuid.UUID
	Name string
}
