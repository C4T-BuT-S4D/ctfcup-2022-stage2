package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"great_mettender/pkg/interfuck"

	"github.com/klauspost/compress/zstd"
)

const (
	// maxLength is 16kb.
	maxLength = 16 * 1024
	opsLimit  = 100000000
)

type executionResult struct {
	Output  string
	Ops     int
	Elapsed time.Duration
	Error   string
}

func main() {
	start := time.Now()
	code := 0
	res, err := run()
	if err != nil {
		res = &executionResult{
			Error: err.Error(),
		}
		code = 1
	}
	res.Elapsed = time.Since(start)
	_ = json.NewEncoder(os.Stdout).Encode(res)
	os.Exit(code)
}

func run() (*executionResult, error) {
	if len(os.Args) < 3 {
		return nil, fmt.Errorf("not enough arguments: %d", len(os.Args))
	}

	programEnc, err := os.ReadFile(os.Args[1])
	if err != nil {
		return nil, fmt.Errorf("reading program file: %w", err)
	}
	inputEnc, err := os.ReadFile(os.Args[2])
	if err != nil {
		return nil, fmt.Errorf("reading input file: %w", err)
	}

	programText, err := decode(string(programEnc))
	if err != nil {
		return nil, fmt.Errorf("decoding program text: %w", err)
	}
	input, err := decode(string(inputEnc))
	if err != nil {
		return nil, fmt.Errorf("decoding input: %w", err)
	}

	program, err := interfuck.Compile(string(programText), opsLimit)

	out, ops, err := program.Run(input)
	if err != nil {
		return nil, fmt.Errorf("running program: %w", err)
	}

	res := &executionResult{
		Output: out,
		Ops:    ops,
	}
	return res, nil
}

// decode if here for you to enjoy reading the traffic.
func decode(s string) ([]byte, error) {
	step1, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decoding step1: %w", err)
	}
	step2, err := zstd.NewReader(bytes.NewReader(step1))
	if err != nil {
		return nil, fmt.Errorf("decoder step2: %w", err)
	}
	step3, err := gzip.NewReader(step2)
	if err != nil {
		return nil, fmt.Errorf("decoder step3: %w", err)
	}

	// Protect our memory.
	res, err := io.ReadAll(io.LimitReader(step3, maxLength))
	if err != nil {
		return nil, fmt.Errorf("reading data: %w", err)
	}
	return res, nil
}
