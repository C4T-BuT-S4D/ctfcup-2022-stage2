package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"gmtchecker/cmd"
	"gmtchecker/lib"
)

func main() {
	var err error
	defer func() {
		var se *lib.StatusError
		if errors.As(err, &se) {
			lib.ProcessStatusError(se)
		}
		lib.ProcessStatusError(lib.NewStatusError(
			lib.VerdictCheckFailed,
			"error in checker",
			fmt.Sprintf("non-status-error encountered: %v", err),
		))
	}()

	if len(os.Args) < 2 {
		err = fmt.Errorf("missing command")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), lib.ActionTimeout)
	defer cancel()

	switch c := os.Args[1]; c {
	case "info":
		err = cmd.Info()
	case "check":
		if len(os.Args) < 3 {
			err = fmt.Errorf("missing host")
			return
		}
		err = cmd.Check(ctx, os.Args[2])
	case "put":
		if len(os.Args) < 6 {
			err = fmt.Errorf("bad args: %d", len(os.Args))
			return
		}
		err = cmd.Put(ctx, os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	case "get":
		if len(os.Args) < 6 {
			err = fmt.Errorf("bad args: %d", len(os.Args))
			return
		}
		err = cmd.Get(ctx, os.Args[2], os.Args[3], os.Args[4], os.Args[5])
	default:
		err = fmt.Errorf("invalid command %s", c)
	}
}
