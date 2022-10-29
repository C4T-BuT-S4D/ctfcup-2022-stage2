package lib

import (
	"fmt"
	"os"
)

type Verdict string

const (
	VerdictOK          Verdict = "OK"
	VerdictMumble      Verdict = "MUMBLE"
	VerdictCorrupt     Verdict = "CORRUPT"
	VerdictDown        Verdict = "DOWN"
	VerdictCheckFailed Verdict = "CHECK FAILED"
)

func (v Verdict) Code() int {
	switch v {
	case VerdictOK:
		return 101
	case VerdictMumble:
		return 102
	case VerdictCorrupt:
		return 103
	case VerdictDown:
		return 104
	default:
		return 110
	}
}

func NewStatusError(verdict Verdict, public, private string) *StatusError {
	return &StatusError{
		verdict: verdict,
		public:  public,
		private: private,
	}
}

type StatusError struct {
	verdict Verdict
	public  string
	private string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("verdict: %v; public: %s; private: %s", e.verdict, e.public, e.private)
}

func ProcessStatusError(err *StatusError) {
	_, _ = fmt.Fprintf(os.Stdout, err.public+"\n")
	_, _ = fmt.Fprintf(os.Stderr, err.private+"\n")
	os.Exit(err.verdict.Code())
}

func Mumble(public, private string, args ...any) *StatusError {
	return NewStatusError(VerdictMumble, public, formatPrivate(public, private, args...))
}

func Corrupt(public, private string, args ...any) *StatusError {
	return NewStatusError(VerdictCorrupt, public, formatPrivate(public, private, args...))
}

func Down(public, private string, args ...any) *StatusError {
	return NewStatusError(VerdictDown, public, formatPrivate(public, private, args...))
}

func OK(public, private string, args ...any) *StatusError {
	return NewStatusError(VerdictOK, public, formatPrivate(public, private, args...))
}

func formatPrivate(public, private string, args ...any) string {
	if private == "" {
		private = public
	} else {
		private = fmt.Sprintf(private, args...)
	}
	return private
}
