package service

type Worker interface {
	Name() string
	Init() error
	Work()
	Destroy()
}

type SimpleWorker struct{}

func (w *SimpleWorker) Name() string { return "SimpleWorker" }

func (w *SimpleWorker) Init() error { return nil }

func (w *SimpleWorker) Work() { panic("not implement") }

func (w *SimpleWorker) Destroy() {}
