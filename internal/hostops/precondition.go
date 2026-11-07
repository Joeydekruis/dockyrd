package hostops

import "errors"

// ErrPrecondition is returned when Plan is blocked (e.g. SSH without key auth).
var ErrPrecondition = errors.New("host action precondition failed")
