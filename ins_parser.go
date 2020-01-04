package pkg

import (
	"fmt"
	"strings"
)

// install instruction
type InsTriple struct {
	First  string
	Second string
	Third  string
}

// todo parse multiple space, and `\"`
func ParseIns(insStr string) (ins InsTriple, err error) {
	insStr = strings.Trim(insStr, " ")
	// parse cmd (the first) of instruction.
	cmdAndLater := strings.SplitN(insStr, " ", 2)
	if len(cmdAndLater) != 2 && len(cmdAndLater) != 1 {
		return ins, fmt.Errorf("syntax error of instruction `%s`", insStr)
	} else {
		ins.First = cmdAndLater[0]
	}

	// first only
	if len(cmdAndLater) == 1 {
		return ins, nil
	}

	// parse second of instruction.
	if strings.HasPrefix(cmdAndLater[1], `"`) {
		// example: `CP "../path" "./des" `
		secondAndLater := strings.SplitN(cmdAndLater[1][1:], `"`, 2)
		// there must at least 2 ", so len must be 2.
		if len(secondAndLater) != 2 {
			return ins, fmt.Errorf("syntax error of instruction `%s`", insStr)
		} else {
			ins.Second = strings.TrimSpace(secondAndLater[0])
			ins.Third = strings.TrimSpace(secondAndLater[1]) // may with `"`
		}
	} else {
		// example: `CP ../path ./des`, `CP ../path`, `CP ../path "./des"`,
		secondAndLater := strings.SplitN(cmdAndLater[1], ` `, 2)
		if len(secondAndLater) == 2 {
			ins.Second = secondAndLater[0]
			ins.Third = secondAndLater[1]
		} else if len(secondAndLater) == 1 {
			ins.Second = secondAndLater[0]
		} else {
			return ins, fmt.Errorf("syntax error of instruction `%s`", insStr)
		}
	}

	// parse third
	ins.Third = strings.Trim(ins.Third, `"`)
	ins.Third = strings.TrimSpace(ins.Third)
	return ins, err
}
