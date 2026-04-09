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
	EventSink   func(RunEvent)
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
	Args       []string

	Cmd   *exec.Cmd
	Stdin io.WriteCloser
	done  chan struct{}

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
	emit := func(event RunEvent) {
		emitRunEvent(options.EventSink, event)
	}

	session := newSession(plan, options)
	session.startedAt = time.Now()
	emit(RunEvent{
		Type:      RunEventStateChanged,
		State:     session.state,
		Reason:    "session created",
		Timestamp: session.startedAt,
	})
	for _, job := range plan.VideoJobs {
		args, err := buildVideoRecordArgs(job, options.MaxDuration)
		if err != nil {
			_ = transitionSession(session, StateFailed, fmt.Sprintf("build video args for %s: %v", job.Source.Name, err), emit)
			stopProcesses(session.processes, gracePeriod)
			return nil, err
		}
		session.markProcessStarted(newManagedProcess(job.Source.Name, args, job.OutputPath))
	}
	for _, job := range plan.AudioJobs {
		args, err := buildAudioMixArgs(job, options.MaxDuration)
		if err != nil {
			_ = transitionSession(session, StateFailed, fmt.Sprintf("build audio args for %s: %v", job.Name, err), emit)
			stopProcesses(session.processes, gracePeriod)
			return nil, err
		}
		session.markProcessStarted(newManagedProcess(job.Name, args, job.OutputPath))
	}
	if len(session.processes) == 0 {
		_ = transitionSession(session, StateFailed, "compiled plan has no executable jobs", emit)
		return nil, errors.New("compiled plan has no executable jobs")
	}

	if err := transitionSession(session, StateRunning, "all workers started", emit); err != nil {
		_ = transitionSession(session, StateFailed, err.Error(), emit)
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
			err := proc.Run(ctx, logger, emit)
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
					if err := beginStopping(session, fmt.Sprintf("%s exited with error: %v", event.Process.Label, event.Err), StateFailed, emit); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
					startStop()
					continue
				}
				if options.MaxDuration > 0 {
					if session.remainingProcesses() == 0 {
						if err := transitionSession(session, StateFinished, "all bounded workers exited cleanly", emit); err != nil {
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
				if err := beginStopping(session, fmt.Sprintf("%s exited before shutdown", event.Process.Label), StateFailed, emit); err != nil {
					cancelProducers()
					_ = group.Wait()
					return nil, err
				}
				startStop()
			case eventCancelRequested:
				if err := beginStopping(session, "cancellation requested: "+event.Reason, StateFinished, emit); err != nil {
					cancelProducers()
					_ = group.Wait()
					return nil, err
				}
				startStop()
			case eventHardTimeout:
				if err := beginStopping(session, event.Reason, StateFailed, emit); err != nil {
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
					if err := transitionSession(session, StateFailed, fmt.Sprintf("graceful stop failed: %v", event.Err), emit); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
				} else if session.finalTarget == StateFailed {
					if err := transitionSession(session, StateFailed, session.reason, emit); err != nil {
						cancelProducers()
						_ = group.Wait()
						return nil, err
					}
				} else {
					if err := transitionSession(session, StateFinished, session.reason, emit); err != nil {
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

func newManagedProcess(label string, args []string, outputPath string) *ManagedProcess {
	return &ManagedProcess{
		Label:      label,
		OutputPath: outputPath,
		Args:       append([]string(nil), args...),
	}
}

func (p *ManagedProcess) Run(ctx context.Context, logger func(string, ...any), emit func(RunEvent)) error {
	if p == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p.OutputPath), 0o755); err != nil {
		return errors.Wrap(err, "create output directory")
	}

	logger("%s argv: ffmpeg %s", p.Label, strings.Join(p.Args, " "))

	cmd := exec.Command("ffmpeg", p.Args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return errors.Wrap(err, "open ffmpeg stdin")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "open ffmpeg stdout")
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return errors.Wrap(err, "open ffmpeg stderr")
	}
	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "start ffmpeg")
	}

	p.mu.Lock()
	if p.done != nil {
		p.mu.Unlock()
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return errors.New("managed process already running")
	}
	p.Cmd = cmd
	p.Stdin = stdin
	p.done = make(chan struct{})
	p.waitErr = nil
	p.mu.Unlock()

	emit(RunEvent{
		Type:         RunEventProcessStarted,
		ProcessLabel: p.Label,
		OutputPath:   p.OutputPath,
	})

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return drainScanner(groupCtx, stderr, func(line string) {
			if strings.TrimSpace(line) != "" {
				logger("%s stderr: %s", p.Label, line)
				emit(RunEvent{
					Type:         RunEventProcessLog,
					ProcessLabel: p.Label,
					OutputPath:   p.OutputPath,
					Stream:       "stderr",
					Message:      line,
				})
			}
		})
	})
	group.Go(func() error {
		return drainScanner(groupCtx, stdout, func(line string) {
			if strings.TrimSpace(line) != "" {
				logger("%s stdout: %s", p.Label, line)
				emit(RunEvent{
					Type:         RunEventProcessLog,
					ProcessLabel: p.Label,
					OutputPath:   p.OutputPath,
					Stream:       "stdout",
					Message:      line,
				})
			}
		})
	})
	group.Go(func() error {
		return cmd.Wait()
	})

	runErr := group.Wait()

	p.mu.Lock()
	p.waitErr = runErr
	p.Cmd = nil
	p.Stdin = nil
	close(p.done)
	p.mu.Unlock()

	return runErr
}

func (p *ManagedProcess) Stop(timeout time.Duration) error {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	stdin := p.Stdin
	cmd := p.Cmd
	done := p.done
	p.mu.RUnlock()

	if stdin != nil {
		_, _ = io.WriteString(stdin, "q\n")
		_ = stdin.Close()
	}
	if done == nil {
		return nil
	}
	select {
	case <-done:
		return p.waitResult()
	case <-time.After(timeout):
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		<-done
		return p.waitResult()
	}
}

func (p *ManagedProcess) Wait() error {
	if p == nil {
		return nil
	}
	p.mu.RLock()
	done := p.done
	p.mu.RUnlock()
	if done == nil {
		return nil
	}
	<-done
	return p.waitResult()
}

func (p *ManagedProcess) waitResult() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.waitErr
}

func drainScanner(ctx context.Context, r io.ReadCloser, fn func(string)) error {
	defer r.Close()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		fn(scanner.Text())
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
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

func transitionSession(session *Session, next SessionState, reason string, emit func(RunEvent)) error {
	if err := session.transition(next, reason); err != nil {
		return err
	}
	emit(RunEvent{
		Type:   RunEventStateChanged,
		State:  session.state,
		Reason: session.reason,
	})
	return nil
}

func beginStopping(session *Session, reason string, finalTarget SessionState, emit func(RunEvent)) error {
	previous := session.state
	if err := session.beginStopping(reason, finalTarget); err != nil {
		return err
	}
	if session.state != previous {
		emit(RunEvent{
			Type:   RunEventStateChanged,
			State:  session.state,
			Reason: session.reason,
		})
	}
	return nil
}
