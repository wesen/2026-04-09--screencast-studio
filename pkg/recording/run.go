package recording

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/wesen/2026-04-09--screencast-studio/pkg/dsl"
)

type RunOptions struct {
	GracePeriod time.Duration
	MaxDuration time.Duration
	Logger      func(string, ...any)
}

type RunResult struct {
	StartedAt  time.Time
	FinishedAt time.Time
	State      SessionState
	Reason     string
	Outputs    []dsl.PlannedOutput
}

type ManagedProcess struct {
	Label      string
	OutputPath string
	Cmd        *exec.Cmd
	Stdin      io.WriteCloser

	done    chan struct{}
	waitErr error
	mu      sync.RWMutex
}

func Run(ctx context.Context, plan *dsl.CompiledPlan, options RunOptions) (*RunResult, error) {
	if plan == nil {
		return nil, errors.New("compiled plan is required")
	}
	if len(plan.VideoJobs) == 0 && len(plan.AudioJobs) == 0 {
		return nil, errors.New("compiled plan has no executable jobs")
	}

	gracePeriod := options.GracePeriod
	if gracePeriod <= 0 {
		gracePeriod = 5 * time.Second
	}
	logger := options.Logger
	if logger == nil {
		logger = func(string, ...any) {}
	}

	session := newSession(plan, options)
	session.startedAt = time.Now()
	for _, job := range plan.VideoJobs {
		args, err := buildVideoRecordArgs(job, options.MaxDuration)
		if err != nil {
			_ = session.transition(StateFailed, fmt.Sprintf("build video args for %s: %v", job.Source.Name, err))
			stopProcesses(session.processes, gracePeriod)
			return nil, err
		}
		proc, err := startManagedProcess(job.Source.Name, args, job.OutputPath, logger)
		if err != nil {
			_ = session.transition(StateFailed, fmt.Sprintf("start process for %s: %v", job.Source.Name, err))
			stopProcesses(session.processes, gracePeriod)
			return nil, err
		}
		session.markProcessStarted(proc)
	}
	for _, job := range plan.AudioJobs {
		args, err := buildAudioMixArgs(job, options.MaxDuration)
		if err != nil {
			_ = session.transition(StateFailed, fmt.Sprintf("build audio args for %s: %v", job.Name, err))
			stopProcesses(session.processes, gracePeriod)
			return nil, err
		}
		proc, err := startManagedProcess(job.Name, args, job.OutputPath, logger)
		if err != nil {
			_ = session.transition(StateFailed, fmt.Sprintf("start process for %s: %v", job.Name, err))
			stopProcesses(session.processes, gracePeriod)
			return nil, err
		}
		session.markProcessStarted(proc)
	}
	if len(session.processes) == 0 {
		_ = session.transition(StateFailed, "compiled plan has no executable jobs")
		return nil, errors.New("compiled plan has no executable jobs")
	}

	if err := session.transition(StateRunning, "all workers started"); err != nil {
		_ = session.transition(StateFailed, err.Error())
		stopProcesses(session.processes, gracePeriod)
		return nil, err
	}

	events := make(chan sessionEvent, len(session.processes)+4)
	producerCtx, cancelProducers := context.WithCancel(context.Background())
	defer cancelProducers()
	group, _ := errgroup.WithContext(producerCtx)

	for _, proc := range session.processes {
		proc := proc
		group.Go(func() error {
			err := proc.Wait()
			publishEvent(producerCtx, events, sessionEvent{
				Type:    eventWorkerExited,
				Process: proc,
				Err:     err,
			})
			return nil
		})
	}

	group.Go(func() error {
		select {
		case <-ctx.Done():
			publishEvent(producerCtx, events, sessionEvent{
				Type:   eventCancelRequested,
				Reason: ctx.Err().Error(),
			})
		case <-producerCtx.Done():
		}
		return nil
	})

	if options.MaxDuration > 0 {
		group.Go(func() error {
			timer := time.NewTimer(boundedRunHardTimeout(options.MaxDuration, gracePeriod))
			defer timer.Stop()
			select {
			case <-timer.C:
				publishEvent(producerCtx, events, sessionEvent{
					Type:   eventHardTimeout,
					Reason: fmt.Sprintf("bounded recording exceeded hard timeout after %s", options.MaxDuration),
				})
			case <-producerCtx.Done():
			}
			return nil
		})
	}

	var (
		stopOnce sync.Once
		stopErr  error
	)
	startStop := func() {
		stopOnce.Do(func() {
			session.stopStarted = true
			group.Go(func() error {
				stopErr = stopProcesses(session.processes, gracePeriod)
				publishEvent(producerCtx, events, sessionEvent{
					Type:   eventStopCompleted,
					Err:    stopErr,
					Reason: session.reason,
				})
				return nil
			})
		})
	}

	for {
		event := <-events
		switch session.state {
		case StateRunning:
			switch event.Type {
			case eventWorkerExited:
				session.markProcessExited(event.Process)
				if event.Err != nil {
					if err := session.beginStopping(fmt.Sprintf("%s exited with error: %v", event.Process.Label, event.Err), StateFailed); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
					startStop()
					continue
				}
				if options.MaxDuration > 0 {
					if session.remainingProcesses() == 0 {
						if err := session.transition(StateFinished, "all bounded workers exited cleanly"); err != nil {
							cancelProducers()
							_ = group.Wait()
							return nil, err
						}
						cancelProducers()
						_ = group.Wait()
						return &RunResult{
							StartedAt:  session.startedAt,
							FinishedAt: session.finishedAt,
							State:      session.state,
							Reason:     session.reason,
							Outputs:    append([]dsl.PlannedOutput(nil), plan.Outputs...),
						}, nil
					}
					continue
				}
				if err := session.beginStopping(fmt.Sprintf("%s exited before shutdown", event.Process.Label), StateFailed); err != nil {
					cancelProducers()
					_ = group.Wait()
					return nil, err
				}
				startStop()
			case eventCancelRequested:
				if err := session.beginStopping("cancellation requested: "+event.Reason, StateFinished); err != nil {
					cancelProducers()
					_ = group.Wait()
					return nil, err
				}
				startStop()
			case eventHardTimeout:
				if err := session.beginStopping(event.Reason, StateFailed); err != nil {
					cancelProducers()
					_ = group.Wait()
					return nil, err
				}
				startStop()
			}
		case StateStopping:
			switch event.Type {
			case eventWorkerExited:
				session.markProcessExited(event.Process)
				if event.Err != nil {
					session.finalTarget = StateFailed
					session.reason = fmt.Sprintf("%s exited during stop with error: %v", event.Process.Label, event.Err)
				}
			case eventStopCompleted:
				if event.Err != nil {
					if err := session.transition(StateFailed, fmt.Sprintf("graceful stop failed: %v", event.Err)); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
				} else if session.finalTarget == StateFailed {
					if err := session.transition(StateFailed, session.reason); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
				} else {
					if err := session.transition(StateFinished, session.reason); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
				}
				cancelProducers()
				_ = group.Wait()
				result := &RunResult{
					StartedAt:  session.startedAt,
					FinishedAt: session.finishedAt,
					State:      session.state,
					Reason:     session.reason,
					Outputs:    append([]dsl.PlannedOutput(nil), plan.Outputs...),
				}
				if session.state == StateFailed {
					return result, errors.New(session.reason)
				}
				return result, nil
			case eventCancelRequested, eventHardTimeout:
				// Ignore duplicate stop requests once stopping has started.
			}
		default:
			cancelProducers()
			_ = group.Wait()
			return nil, fmt.Errorf("unexpected terminal session state %s", session.state)
		}
	}
}

func stopProcesses(processes []*ManagedProcess, timeout time.Duration) error {
	if len(processes) == 0 {
		return nil
	}
	errs := []string{}
	for _, proc := range processes {
		if err := proc.Stop(timeout); err != nil {
			errs = append(errs, errors.Wrapf(err, "stop %s", proc.Label).Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, "; "))
}

func startManagedProcess(label string, args []string, outputPath string, logger func(string, ...any)) (*ManagedProcess, error) {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return nil, errors.Wrap(err, "create output directory")
	}

	logger("%s argv: ffmpeg %s", label, strings.Join(args, " "))

	cmd := exec.Command("ffmpeg", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, errors.Wrap(err, "open ffmpeg stdin")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "open ffmpeg stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "open ffmpeg stderr")
	}
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "start ffmpeg")
	}

	proc := &ManagedProcess{
		Label:      label,
		OutputPath: outputPath,
		Cmd:        cmd,
		Stdin:      stdin,
		done:       make(chan struct{}),
	}
	go drainScanner(stderr, func(line string) {
		if strings.TrimSpace(line) != "" {
			logger("%s stderr: %s", label, line)
		}
	})
	go drainScanner(stdout, func(line string) {
		if strings.TrimSpace(line) != "" {
			logger("%s stdout: %s", label, line)
		}
	})
	go func() {
		proc.mu.Lock()
		proc.waitErr = cmd.Wait()
		proc.mu.Unlock()
		close(proc.done)
	}()
	return proc, nil
}

func (p *ManagedProcess) Stop(timeout time.Duration) error {
	if p == nil {
		return nil
	}
	if p.Stdin != nil {
		_, _ = io.WriteString(p.Stdin, "q\n")
		_ = p.Stdin.Close()
	}
	select {
	case <-p.done:
		return p.waitResult()
	case <-time.After(timeout):
		if p.Cmd != nil && p.Cmd.Process != nil {
			_ = p.Cmd.Process.Kill()
		}
		<-p.done
		return p.waitResult()
	}
}

func (p *ManagedProcess) Wait() error {
	if p == nil {
		return nil
	}
	<-p.done
	return p.waitResult()
}

func (p *ManagedProcess) waitResult() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.waitErr
}

func drainScanner(r io.Reader, fn func(string)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fn(scanner.Text())
	}
}

func publishEvent(ctx context.Context, events chan<- sessionEvent, event sessionEvent) {
	select {
	case <-ctx.Done():
	case events <- event:
	}
}

func boundedRunHardTimeout(maxDuration, gracePeriod time.Duration) time.Duration {
	timeout := maxDuration + gracePeriod
	if timeout <= 0 {
		return gracePeriod
	}
	return timeout
}
