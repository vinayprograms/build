// Package output provides build output formatting and reporting.
//
// This package implements various output reporters for displaying build progress
// and results to users. The primary reporter is NormalReporter which outputs:
//
//   - Target being built
//   - Command output (stdout/stderr)
//   - Completion/failure status
//
// Example usage:
//
//	reporter := output.NewNormalReporter(os.Stdout)
//	reporter.BuildStarted("build/app")
//	reporter.CommandOutput("gcc -c main.c", "", "")
//	reporter.BuildCompleted("build/app", true, "")
//	reporter.Summary(1, 0)
package output
