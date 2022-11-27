package checker

import (
	"github.com/pomo-mondreganto/go-checklib"
)

type Checker struct{}

func (ch *Checker) Info() *checklib.CheckerInfo {
	return &checklib.CheckerInfo{
		Vulns:      1,
		Timeout:    10,
		AttackData: true,
		Puts:       3,
		Gets:       10,
	}
}
