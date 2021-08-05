package stages

import "time"

// SleepTransportUnavailable is used to determine how long to sleep between deliver attempts
const SleepTransportUnavailable = 3 * time.Second
