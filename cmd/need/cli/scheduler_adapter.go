package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vinayprograms/need/internal/eval"
	"github.com/vinayprograms/need/internal/executor"
	"github.com/vinayprograms/need/internal/planner"
)

// GetParallelDirective extracts the .parallel: value from a parsed Needfile.
// Returns 0 if the directive is not set or invalid.
func GetParallelDirective(result NeedfileResult) int {
	for _, stmt := range result.Statements() {
		if stmt.StatementType() != "directive" {
			continue
		}
		summary := stmt.Summary()
		if !strings.HasPrefix(summary, ".parallel:") {
			continue
		}
		// Extract the value after ".parallel: "
		valueStr := strings.TrimSpace(strings.TrimPrefix(summary, ".parallel:"))
		value, err := strconv.Atoi(valueStr)
		if err != nil {
			return 0
		}
		return value
	}
	return 0
}

// ResolveWorkerCount determines the number of workers to use based on CLI -j flag
// and .parallel: directive value. The -j flag takes precedence when explicitly set
// (greater than 1), otherwise the .parallel: directive value is used.
//
// This re-exports executor.ResolveWorkerCount for CLI use.
func ResolveWorkerCount(cliJobs, parallelDirective int) int {
	return executor.ResolveWorkerCount(cliJobs, parallelDirective)
}

// SchedulerResult represents the result of a single task execution.
type SchedulerResult struct {
	Target  string
	Results []*ExecResult
	Error   error
	Skipped bool
}

// Scheduler wraps executor.Scheduler for CLI use.
type Scheduler struct {
	sched *executor.Scheduler
}

// NewScheduler creates a scheduler with the given number of workers.
func NewScheduler(exec *Executor, numWorkers int) *Scheduler {
	return &Scheduler{
		sched: executor.NewScheduler(exec.exec, numWorkers),
	}
}

// SetKeepGoing enables or disables keep-going mode.
func (s *Scheduler) SetKeepGoing(keepGoing bool) {
	s.sched.SetKeepGoing(keepGoing)
}

// Workers returns the number of workers.
func (s *Scheduler) Workers() int {
	return s.sched.Workers()
}

// ExecutePlan executes a build plan using the scheduler.
// Returns results for all tasks in the plan.
func (s *Scheduler) ExecutePlan(
	planResult BuildPlanResult,
	ctx EvalContext,
	reporter OutputReporter,
) []SchedulerResult {
	// Get the underlying plan
	adapter, ok := planResult.(*buildPlanResultAdapter)
	if !ok || adapter.plan == nil {
		return nil
	}

	// Get the underlying eval context
	eca, ok := ctx.(*evalContextAdapter)
	if !ok {
		return nil
	}

	// Create context factory for the scheduler
	ctxFactory := func(target string) *eval.CommandContext {
		// Find the task to get its dependencies
		var deps []string
		for _, task := range adapter.plan.Tasks {
			if task.Target == target {
				deps = task.Dependencies
				break
			}
		}
		return eval.NewCommandContext(eca.ctx, target, deps)
	}

	// Create callback for build started events
	callback := func(target string) {
		reporter.BuildStarted(target)
	}

	// Execute using the scheduler
	results := s.sched.ExecuteWithCallback(adapter.plan.Tasks, ctxFactory, callback)

	// Convert results
	schedulerResults := make([]SchedulerResult, len(results))
	for i, r := range results {
		var execResults []*ExecResult
		for _, er := range r.Results {
			execResults = append(execResults, &ExecResult{result: er})
		}
		schedulerResults[i] = SchedulerResult{
			Target:  r.Target,
			Results: execResults,
			Error:   r.Error,
			Skipped: r.Skipped,
		}
	}

	return schedulerResults
}

// GetPlanTasks returns the underlying tasks from a BuildPlanResult.
// This is used to access task details for reporting.
func GetPlanTasks(planResult BuildPlanResult) []planner.BuildTask {
	adapter, ok := planResult.(*buildPlanResultAdapter)
	if !ok || adapter.plan == nil {
		return nil
	}
	return adapter.plan.Tasks
}

// TaskError represents a task execution error.
type TaskError struct {
	Target   string
	ExitCode int
	Command  string
}

func (e *TaskError) Error() string {
	return fmt.Sprintf("command failed with exit code %d: %s", e.ExitCode, e.Command)
}
