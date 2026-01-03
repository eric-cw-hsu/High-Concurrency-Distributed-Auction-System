package service

import "time"

type SessionPolicy struct {
	MaxSessions int
	TTL         time.Duration
}

func (p SessionPolicy) CanCreate(current int) bool {
	return current < p.MaxSessions
}
