package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"great_mettender/internal/models"
)

const timeout = time.Second * 10

func NewExecutor(path string) *Executor {
	return &Executor{path: path}
}

type Executor struct {
	path string
}

func (e *Executor) Execute(ctx context.Context, program, input string) (*models.ExecutionResult, error) {
	programFile, err := os.CreateTemp("", "program")
	if err != nil {
		return nil, fmt.Errorf("creating temp for program: %w", err)
	}
	defer cleanupTemp(programFile)

	inputFile, err := os.CreateTemp("", "input")
	if err != nil {
		return nil, fmt.Errorf("creating temp for input: %w", err)
	}
	defer cleanupTemp(inputFile)

	if _, err := programFile.WriteString(program); err != nil {
		return nil, fmt.Errorf("writing program: %w", err)
	}
	if _, err := inputFile.WriteString(input); err != nil {
		return nil, fmt.Errorf("writing input: %w", err)
	}

	stdout := bytes.Buffer{}
	start := time.Now()
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(rctx, e.path, programFile.Name(), inputFile.Name())
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		if errors.Is(err, context.Canceled) {
			// Timeout.
			return &models.ExecutionResult{
				Elapsed: time.Since(start),
				Error:   "command canceled early",
			}, nil
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				var res models.ExecutionResult
				if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
					return &models.ExecutionResult{
						Elapsed: time.Since(start),
						Error:   fmt.Sprintf("unable to parse command output: %s", stdout.String()),
					}, nil
				}
				return &res, nil
			} else {
				// Unexpected return code, should never happen.
				// Panic?
				return &models.ExecutionResult{
					Elapsed: time.Since(start),
					Error: fmt.Sprintf(
						"[unexpected1] running program %s input %s: %s; %s",
						program,
						input,
						exitErr.String(),
						string(exitErr.Stderr),
					),
				}, nil
			}
		}
		// Even more unexpected error.
		return &models.ExecutionResult{
			Elapsed: time.Since(start),
			Error: fmt.Sprintf(
				"[unexpected2] running program %s input %s: %v",
				program,
				input,
				err,
			),
		}, nil
	}
	var res models.ExecutionResult
	if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
		return &models.ExecutionResult{
			Elapsed: time.Since(start),
			Error:   fmt.Sprintf("unable to parse command output: %s", stdout.String()),
		}, nil
	}
	return &res, nil
}

func cleanupTemp(f *os.File) {
	_ = f.Close()
	_ = os.Remove(f.Name())
}
