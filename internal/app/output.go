package app

import (
	"fmt"

	"github.com/pterm/pterm"
)

type OutputMode int

const (
	MachineOutputMode OutputMode = iota
	HumanOutputMode
)

func (a Application) outputSection(message string, args ...interface{}) {
	if a.outputMode == MachineOutputMode {
		return
	}

	pterm.DefaultSection.Println(fmt.Sprintf(message, args...))
}

func (a Application) outputInfo(message string, args ...interface{}) {
	if a.outputMode == MachineOutputMode {
		return
	}

	pterm.Info.Printfln(message, args...)
}

func (a Application) outputSuccess(message string, args ...interface{}) {
	if a.outputMode == MachineOutputMode {
		return
	}

	pterm.Success.Printfln(message, args...)
}

func (a Application) outputCreateProgressBar() func(message string, total int) func() {
	return func(message string, total int) func() {
		if a.outputMode == MachineOutputMode {
			return func() {}
		}

		p, _ := pterm.DefaultProgressbar.WithTotal(total).WithTitle(message).Start()

		return func() {
			p.Increment()
		}
	}
}
