package util

import (
	"context"
	"sync"
	"sync/atomic"
)

const (
	RecordStatNew     uint32 = 1 // 新建状态
	RecordStatInit    uint32 = 2 // 初始化状态
	RecordStatRunning uint32 = 3 // running状态
	RecordStatClosed  uint32 = 4 // 关闭状态
)

type Recorder struct {
	conf *RecorderConfig

	stat   uint32
	cache  chan *FileInfo
	cancel context.CancelFunc
}

func NewRecorder(conf *RecorderConfig) *Recorder {
	return &Recorder{
		conf:  conf,
		stat:  RecordStatNew,
		cache: make(chan *FileInfo, conf.Capa),
	}
}

func (r *Recorder) Init() {
	if !atomic.CompareAndSwapUint32(&r.stat, RecordStatNew, RecordStatInit) {
		return
	}
	// todo
	Log().Infof("recorder init success")
}

func (r *Recorder) Run(wg *sync.WaitGroup) {
	if !atomic.CompareAndSwapUint32(&r.stat, RecordStatInit, RecordStatRunning) {
		return
	}

	wg.Add(1)
	go func() {
		defer func() {
			close(r.cache) // 关闭管道
			Log().Infof("stop run recorder success")

			wg.Done()
		}()

		var ctx context.Context
		ctx, r.cancel = context.WithCancel(context.Background())
		for {
			select {
			case <-ctx.Done():
				for len(r.cache) > 0 {
					r.doSave(<-r.cache)
				}
				return
			case f := <-r.cache:
				if err := r.doSave(f); err != nil {
					Log().Errorf("save record:%v fail, err:%v", f.filePathName, err)
					continue
				}
			}
		}
	}()
	Log().Infof("run recorder success")
}

func (r *Recorder) doSave(f *FileInfo) error {
	return f.Set(context.Background())
}

func (r *Recorder) Close() {
	if !atomic.CompareAndSwapUint32(&r.stat, RecordStatRunning, RecordStatClosed) {
		return
	}
	r.cancel()
}

func (r *Recorder) Add(f *FileInfo) {
	r.cache <- f
}

func (r *Recorder) Get(ctx context.Context, f *FileInfo) error {
	return f.Get(ctx)
}
