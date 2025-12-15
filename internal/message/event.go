package message

import "time"

type Event interface {
	OccurredOn() time.Time
	EventType() string
	EventName() string
	GetAggregateID() string
}
