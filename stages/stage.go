package stages

import (
	"sync"

	"github.com/2gis/loggo/logging"
)

type stage struct {
	proceed func()
	wg      *sync.WaitGroup
	logger  logging.Logger
}

func (s *stage) Close() {
	s.wg.Wait()
}

// InitWorker starts stage worker
func (s *stage) InitWorker() {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		s.proceed()
	}()
}

// StageInit initializes parallelism count of stage workers, then waits for stage to close
func StageInit(stage Stage, parallelism int) {
	for i := 0; i < parallelism; i++ {
		stage.InitWorker()
	}

	stage.Close()
}
