package client

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/pomo-mondreganto/go-checklib/gen"
)

//go:embed programs/bottles1.txt
var bottles1 string

//go:embed programs/bottles2.txt
var bottles2 string

//go:embed programs/bottles3.txt
var bottles3 string

//go:embed programs/hello1.txt
var hello1 string

//go:embed programs/hello2.txt
var hello2 string

//go:embed programs/cat1.txt
var cat1 string

//go:embed programs/cat2.txt
var cat2 string

//go:embed programs/cellsize.txt
var cellsize string

//go:embed programs/triag.txt
var triag string

var allPrograms = []string{
	bottles1,
	bottles2,
	bottles3,
	hello1,
	hello2,
	cat1,
	cat2,
	cellsize,
	triag,
}

var knownOutputPrograms = []string{
	hello1,
	hello2,
	cat1,
	cat2,
	cellsize,
}

var catPrograms = []string{
	cat1,
	cat2,
}

type Program struct {
	Text           string
	ExpectedOutput func(input string) string
}

func EncodeFormat(s string) string {
	buf := bytes.Buffer{}
	// program -> gzip -> zstd -> base64.
	step2, _ := zstd.NewWriter(&buf)
	step1 := gzip.NewWriter(step2)
	_, _ = step1.Write([]byte(s))
	_ = step1.Close()
	_ = step2.Close()
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func DecodeFormat(s string) (string, error) {
	step1, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("decoding step1: %w", err)
	}
	step2, err := zstd.NewReader(bytes.NewReader(step1))
	if err != nil {
		return "", fmt.Errorf("decoder step2: %w", err)
	}
	step3, err := gzip.NewReader(step2)
	if err != nil {
		return "", fmt.Errorf("decoder step3: %w", err)
	}

	res, err := io.ReadAll(io.LimitReader(step3, 1024))
	if err != nil {
		return "", fmt.Errorf("reading data: %w", err)
	}
	return string(res), nil
}

func mutateProgram(s string) string {
	left := []rune(s)
	for i := 0; i*2 < len(left); i++ {
		left[i], left[len(left)-i-1] = left[len(left)-i-1], left[i]
	}

	b := strings.Builder{}
	for len(left) > 0 {
		// Inject random NOPs.
		if b.Len()+len(left) < 50000 && gen.RandInt(0, 50) == 0 {
			b.WriteRune('>')
			left = append(left, '<')
		}
		if b.Len()+len(left) < 50000 && gen.RandInt(0, 50) == 0 {
			b.WriteRune('+')
			left = append(left, '-')
		}
		if b.Len()+len(left) < 50000 && gen.RandInt(0, 50) == 0 {
			b.WriteRune('-')
			left = append(left, '+')
		}

		b.WriteRune(left[len(left)-1])
		left = left[:len(left)-1]
	}
	return b.String()
}

func SampleProgram() string {
	program := gen.Sample(allPrograms)
	program = hello2
	return EncodeFormat(mutateProgram(program))
}

func SampleProgramWithOutput(input string) (string, string) {
	program := gen.Sample(knownOutputPrograms)
	var out string
	switch program {
	case hello1, hello2:
		out = "Hello World!\n"
	case cat1, cat2:
		out = input
	case cellsize:
		out = "16 bit cells"
	}
	return EncodeFormat(mutateProgram(program)), out
}

func SampleCatProgram() string {
	program := gen.Sample(catPrograms)
	return EncodeFormat(mutateProgram(program))
}
