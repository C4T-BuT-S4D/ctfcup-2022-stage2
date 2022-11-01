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

	err = cmd.Run()

	var exitErr *exec.ExitError
	switch {
	case err == nil:
		var res models.ExecutionResult
		if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
			return nil, fmt.Errorf("parsing interpreter result: %w", err)
		}
		return &res, nil

	case errors.Is(err, context.Canceled):
		// Timeout.
		return &models.ExecutionResult{
			Elapsed: time.Since(start),
			Error:   "command canceled early",
		}, nil

	case errors.As(err, &exitErr):
		switch code := exitErr.ExitCode(); code {
		case 1:
			// os.Exit(1) in interpreter.
			var res models.ExecutionResult
			if err := json.Unmarshal(stdout.Bytes(), &res); err != nil {
				return &models.ExecutionResult{
					Elapsed: time.Since(start),
					Error:   fmt.Sprintf("unable to parse command output: %s", stdout.String()),
				}, nil
			}
			return &res, nil

		default:
			return nil, fmt.Errorf("unexpected process error (code %d): %w", code, err)
		}

	default:
		return nil, fmt.Errorf("unexpected cmd error: %w", err)
	}
}

func cleanupTemp(f *os.File) {
	_ = f.Close()
	_ = os.Remove(f.Name())
}
