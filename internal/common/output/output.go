package output

import (
	"fmt"
	"log/slog"

	"github.com/pterm/pterm"
)

type Mode int

const (
	MachineMode Mode = iota
	HumanMode
)

type Output struct {
	mode Mode
}

func New(mode Mode) Output {
	return Output{mode: mode}
}

func (o Output) Println(message string) {
	pterm.Println(message)
}

func (o Output) Section(message string, args ...interface{}) {
	if o.mode == MachineMode {
		return
	}

	pterm.DefaultSection.Println(fmt.Sprintf(message, args...))
}

func (o Output) Info(message string, args ...interface{}) {
	if o.mode == MachineMode {
		return
	}

	pterm.Info.Printfln(message, args...)
}

func (o Output) Success(message string, args ...interface{}) {
	if o.mode == MachineMode {
		return
	}

	pterm.Success.Printfln(message, args...)
}

type ProgressBarCreator func(message string, total int) func()

func (o Output) NewProgressBarCreator() ProgressBarCreator {
	return func(message string, total int) func() {
		if o.mode == MachineMode {
			return func() {}
		}

		p, err := pterm.DefaultProgressbar.WithTotal(total).WithTitle(message).Start()
		if err != nil {
			slog.Error("error creating progress bar", slog.Any("error", err))
			return func() {}
		}

		return func() {
			p.Increment()
		}
	}
}

func (o Output) Spinner(startMessage, doneMessage string) func() {
	spinnerInfo, _ := pterm.DefaultSpinner.Start(startMessage)
	return func() {
		spinnerInfo.Success(doneMessage)
	}
}

func (o Output) Table(tableData [][]string) {
	err := pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	if err != nil {
		slog.Error("error rendering table", slog.Any("error", err))
	}
}
