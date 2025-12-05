// Package output is a thin wrapper around pterm that noo-ps if the app is running
// in machine mode. All non log CLI messages intended for human interactive mode
// should use this package.
package output

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/pterm/pterm"
)

type Mode int

const (
	MachineMode Mode = iota
	HumanMode
)

var (
	_ Output = (*humanOutput)(nil)
	_ Output = (*noopOutput)(nil)
)

type Output interface {
	Println(message string)
	Section(message string, args ...interface{})
	Info(message string, args ...interface{})
	Success(message string, args ...interface{})
	NewProgressBarCreator() ProgressBarCreator
	Spinner(startMessage, doneMessage string) func()
	Table(tableData [][]string)
}

func New(mode Mode) Output {
	if mode == MachineMode {
		return noopOutput{}
	}
	return humanOutput{}
}

type humanOutput struct{}

func (o humanOutput) Println(message string) {

	pterm.Println(message)
}

func (o humanOutput) Section(message string, args ...interface{}) {
	pterm.DefaultSection.Println(fmt.Sprintf(message, args...))
}

func (o humanOutput) Info(message string, args ...interface{}) {
	pterm.Info.Printfln(message, args...)
}

func (o humanOutput) Success(message string, args ...interface{}) {
	pterm.Success.Printfln(message, args...)
}

type ProgressBarCreator func(message string, total int) func()

func (o humanOutput) NewProgressBarCreator() ProgressBarCreator {
	return func(message string, total int) func() {
		p, err := pterm.DefaultProgressbar.WithTotal(total).WithTitle(message).Start()
		if err != nil {
			slog.Error("error creating progress bar", slog.Any("error", err))
			return func() {}
		}

		mu := sync.Mutex{}
		return func() {
			mu.Lock()
			p.Increment()
			mu.Unlock()
		}
	}
}

func (o humanOutput) Spinner(startMessage, doneMessage string) func() {
	spinnerInfo, _ := pterm.DefaultSpinner.Start(startMessage)
	return func() {
		spinnerInfo.Success(doneMessage)
	}
}

func (o humanOutput) Table(tableData [][]string) {
	err := pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(tableData).Render()
	if err != nil {
		slog.Error("error rendering table", slog.Any("error", err))
	}
}

type noopOutput struct{}

func (n noopOutput) Println(_ string) {
	return
}

func (n noopOutput) Section(_ string, _ ...interface{}) {
	return
}

func (n noopOutput) Info(_ string, _ ...interface{}) {
	return
}

func (n noopOutput) Success(_ string, _ ...interface{}) {
	return
}

func (n noopOutput) NewProgressBarCreator() ProgressBarCreator {
	return func(_ string, _ int) func() {
		return func() {}
	}
}

func (n noopOutput) Spinner(_, _ string) func() {
	return func() {}
}

func (n noopOutput) Table(_ [][]string) {
}
