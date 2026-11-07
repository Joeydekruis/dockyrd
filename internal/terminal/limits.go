package terminal

import "time"

const (
	TicketTTL            = 60 * time.Second
	IdleTimeout          = 15 * time.Minute
	MaxSessionDuration   = 2 * time.Hour
	MaxConcurrentPerUser = 3
)
