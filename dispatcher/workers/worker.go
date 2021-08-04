package workers

import (
	"errors"
	"sync"

	"github.com/2gis/loggo/logging"
)

var ErrNoRecords = errors.New("no data read")

type worker struct {
	sync.Mutex
	wg     *sync.WaitGroup
	logger logging.Logger
}
