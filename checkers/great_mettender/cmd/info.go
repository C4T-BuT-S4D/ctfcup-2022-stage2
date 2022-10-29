package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"gmtchecker/lib"
)

func Info() error {
	info := lib.CheckerInfo{
		Vulns:      1,
		Timeout:    10,
		AttackData: true,
	}
	if err := json.NewEncoder(os.Stdout).Encode(info); err != nil {
		return fmt.Errorf("encoding info: %w", err)
	}
	return lib.OK("", "")
}
