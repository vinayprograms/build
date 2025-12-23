package executor

import (
	"sync"

	"github.com/vinayprograms/build/internal/eval"
	"github.com/vinayprograms/build/internal/planner"
)

// TaskResult represents the result of executing a single build task.
type TaskResult struct {
	Target  string        // The target that was built
	Results []*ExecResult // Results from recipe execution
	Error   error         // Error if execution failed
	Skipped bool          // True if task was skipped due to dependency failure
}

// ContextFactory creates a CommandContext for a target.
type ContextFactory func(target string) *eval.CommandContext

// TaskCallback is called when a task starts execution.
type TaskCallback func(target string)

// Scheduler handles parallel execution of build tasks.
type Scheduler struct {
	executor   *Executor
	numWorkers int
	keepGoing  bool // If true, continue building after failures
}

// NewScheduler creates a scheduler with the given number of workers.
func NewScheduler(executor *Executor, numWorkers int) *Scheduler {
	if numWorkers < 1 {
		numWorkers = 1
	}
	return &Scheduler{
		executor:   executor,
		numWorkers: numWorkers,
	}
}

// ResolveWorkerCount determines the number of workers to use based on CLI -j flag
// and .parallel: directive value. The -j flag takes precedence when explicitly set
// (greater than 1), otherwise the .parallel: directive value is used.
//
// Parameters:
//   - cliJobs: value from -j flag (default is 1)
//   - parallelDirective: value from .parallel: directive in Buildfile (0 if not set)
//
// Returns the number of workers to use, minimum 1.
func ResolveWorkerCount(cliJobs, parallelDirective int) int {
	// CLI flag explicitly set to value > 1 takes precedence
	if cliJobs > 1 {
		return cliJobs
	}

	// If CLI is default (1) or invalid, use directive if valid
	if parallelDirective > 0 {
		return parallelDirective
	}

	// Default to 1 if nothing valid provided
	if cliJobs > 0 {
		return cliJobs
	}
	return 1
}

// SetKeepGoing enables or disables keep-going mode.
// When enabled, the scheduler continues building independent targets after a failure.
// Dependent targets are still skipped.
func (s *Scheduler) SetKeepGoing(keepGoing bool) {
	s.keepGoing = keepGoing
}

// Workers returns the number of workers.
func (s *Scheduler) Workers() int {
	return s.numWorkers
}

// Execute executes all tasks respecting dependencies.
func (s *Scheduler) Execute(tasks []planner.BuildTask, ctxFactory ContextFactory) []TaskResult {
	return s.ExecuteWithCallback(tasks, ctxFactory, nil)
}

// ExecuteWithCallback executes tasks with a callback called at task start.
func (s *Scheduler) ExecuteWithCallback(tasks []planner.BuildTask, ctxFactory ContextFactory, callback TaskCallback) []TaskResult {
	if len(tasks) == 0 {
		return nil
	}

	// Build task map and dependency tracking
	taskMap := make(map[string]*planner.BuildTask)
	for i := range tasks {
		taskMap[tasks[i].Target] = &tasks[i]
	}

	// Track results
	results := make([]TaskResult, len(tasks))
	resultMap := make(map[string]int) // target -> index in results
	for i, t := range tasks {
		resultMap[t.Target] = i
	}

	// Track completed and failed tasks
	var mu sync.Mutex
	completed := make(map[string]bool)
	failed := make(map[string]bool)
	cancelled := false

	// Channel for tasks ready to execute
	ready := make(chan *planner.BuildTask, len(tasks))
	done := make(chan string, len(tasks))

	// Count pending dependencies for each task
	pendingDeps := make(map[string]int)
	for _, t := range tasks {
		count := 0
		for _, dep := range t.Dependencies {
			if _, ok := taskMap[dep]; ok {
				count++
			}
		}
		pendingDeps[t.Target] = count
	}

	// Queue tasks with no dependencies
	for _, t := range tasks {
		if pendingDeps[t.Target] == 0 {
			ready <- &tasks[resultMap[t.Target]]
		}
	}

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < s.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range ready {
				mu.Lock()
				if cancelled {
					// Mark as skipped
					idx := resultMap[task.Target]
					results[idx] = TaskResult{
						Target:  task.Target,
						Skipped: true,
					}
					mu.Unlock()
					done <- task.Target
					continue
				}

				// Check if any dependency failed
				depFailed := false
				for _, dep := range task.Dependencies {
					if failed[dep] {
						depFailed = true
						break
					}
				}
				mu.Unlock()

				if depFailed {
					mu.Lock()
					idx := resultMap[task.Target]
					results[idx] = TaskResult{
						Target:  task.Target,
						Skipped: true,
					}
					failed[task.Target] = true
					mu.Unlock()
					done <- task.Target
					continue
				}

				// Execute callback
				if callback != nil {
					callback(task.Target)
				}

				// Execute the task
				result := s.executeTask(task, ctxFactory)

				mu.Lock()
				idx := resultMap[task.Target]
				results[idx] = result

				if result.Error != nil {
					failed[task.Target] = true
					if !s.keepGoing {
						cancelled = true // Cancel remaining tasks only if not in keep-going mode
					}
				} else {
					completed[task.Target] = true
				}
				mu.Unlock()

				done <- task.Target
			}
		}()
	}

	// Monitor completions and queue newly ready tasks
	go func() {
		remaining := len(tasks)
		for remaining > 0 {
			target := <-done
			remaining--

			mu.Lock()
			// Check which tasks are now ready
			for _, t := range tasks {
				if completed[t.Target] || failed[t.Target] {
					continue
				}
				if pendingDeps[t.Target] == 0 {
					continue // Already queued
				}

				// Decrement pending count if this was a dependency
				for _, dep := range t.Dependencies {
					if dep == target {
						pendingDeps[t.Target]--
						if pendingDeps[t.Target] == 0 {
							ready <- taskMap[t.Target]
						}
						break
					}
				}
			}
			mu.Unlock()
		}
		close(ready)
	}()

	wg.Wait()
	return results
}

// executeTask executes a single task.
func (s *Scheduler) executeTask(task *planner.BuildTask, ctxFactory ContextFactory) TaskResult {
	result := TaskResult{
		Target: task.Target,
	}

	// If no recipe, nothing to execute
	if task.Recipe == nil {
		return result
	}

	// Create context
	cmdCtx := ctxFactory(task.Target)
	if task.Captures != nil {
		cmdCtx.SetCaptures(task.Captures)
		// Set stem to the first capture value (similar to make's $*)
		// For pattern targets like src/{name}.txt matching src/input1.txt,
		// stem would be "input1"
		for _, v := range task.Captures {
			cmdCtx.SetStem(v)
			break
		}
	}

	// Execute recipe
	results, err := s.executor.ExecuteRecipe(task.Recipe, cmdCtx)
	result.Results = results
	result.Error = err

	return result
}
