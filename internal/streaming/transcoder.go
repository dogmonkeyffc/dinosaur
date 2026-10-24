// Package streaming adapts local files and FFmpeg output for browser playback.
package streaming

import (
	"context"
	"errors"
	"io"
	"os/exec"
	"sync"
	"time"

	"dinosaur/internal/playback"
)

var ErrBusy = errors.New("playback capacity exhausted")

type Transcoder struct {
	executable string
	slots      chan struct{}
	ctx        context.Context
	cancel     context.CancelFunc
	mu         sync.Mutex
	closed     bool
	workers    sync.WaitGroup
}

func New(executable string) *Transcoder {
	if executable == "" {
		executable = "ffmpeg"
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Transcoder{executable: executable, slots: make(chan struct{}, 2), ctx: ctx, cancel: cancel}
}

type Stream struct {
	io.ReadCloser
	command     *exec.Cmd
	cancel      context.CancelFunc
	stopParent  func() bool
	release     func()
	once        sync.Once
	startup     *time.Timer
	diagnostics *stderrTail
}

func (stream *Stream) Close() error {
	stream.once.Do(func() {
		stream.startup.Stop()
		stream.cancel()
		stream.stopParent()
		_ = stream.ReadCloser.Close()
		_ = stream.command.Wait()
		stream.release()
	})
	return nil
}
func (t *Transcoder) Open(ctx context.Context, path string, plan playback.Plan) (*Stream, error) {
	waiting, cancelWaiting := context.WithTimeout(ctx, 5*time.Second)
	defer cancelWaiting()
	select {
	case t.slots <- struct{}{}:
	case <-waiting.Done():
		return nil, ErrBusy
	case <-t.ctx.Done():
		return nil, context.Canceled
	}
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		<-t.slots
		return nil, context.Canceled
	}
	t.workers.Add(1)
	t.mu.Unlock()
	release := func() { <-t.slots; t.workers.Done() }
	worker, cancel := context.WithTimeout(ctx, 6*time.Hour)
	stop := context.AfterFunc(t.ctx, cancel)
	command := exec.CommandContext(worker, t.executable, Arguments(path, plan)...)
	command.WaitDelay = time.Second
	diagnostics := &stderrTail{}
	command.Stderr = diagnostics
	output, err := command.StdoutPipe()
	if err == nil {
		err = command.Start()
	}
	if err != nil {
		cancel()
		stop()
		if output != nil {
			_ = output.Close()
		}
		release()
		return nil, err
	}
	stream := &Stream{ReadCloser: output, command: command, cancel: cancel, stopParent: stop, release: release, diagnostics: diagnostics}
	stream.startup = time.AfterFunc(20*time.Second, cancel)
	return stream, nil
}
func (t *Transcoder) Close() {
	t.mu.Lock()
	t.closed = true
	t.cancel()
	t.mu.Unlock()
	t.workers.Wait()
}

func (stream *Stream) Ready() { stream.startup.Stop() }
