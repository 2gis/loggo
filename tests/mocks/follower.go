package mocks

import "context"

type WorkerFollowerMock struct {
	activeFlag bool
}

func (w *WorkerFollowerMock) Start(ctx context.Context) {
	w.activeFlag = true
	<-ctx.Done()
	w.activeFlag = false
}

func (w *WorkerFollowerMock) SetEOFShutdownFlag() {}

func (w *WorkerFollowerMock) Stop() {
	w.activeFlag = false
}

func (w *WorkerFollowerMock) GetActiveFlag() bool {
	return w.activeFlag
}

type WorkerJournaldMock struct {
}

func (w *WorkerJournaldMock) Start(ctx context.Context) {
	<-ctx.Done()
}
