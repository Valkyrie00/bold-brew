package services

import (
	"bbrew/internal/models"
	"fmt"
	"github.com/rivo/tview"
	"io"
	"os/exec"
	"sync"
)

type CommandServiceInterface interface {
	UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error
	UpdatePackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error
	RemovePackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error
	InstallPackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error
}

type CommandService struct{}

var NewCommandService = func() CommandServiceInterface {
	return &CommandService{}
}

func (s *CommandService) UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "upgrade") // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

func (s *CommandService) UpdatePackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "upgrade", info.Name) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

func (s *CommandService) RemovePackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "remove", info.Name) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

func (s *CommandService) InstallPackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "install", info.Name) // #nosec G204
	return s.executeCommand(app, cmd, outputView)
}

func (s *CommandService) executeCommand(
	app *tview.Application,
	cmd *exec.Cmd,
	outputView *tview.TextView,
) error {
	stdoutPipe, stdoutWriter := io.Pipe()
	stderrPipe, stderrWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	if err := cmd.Start(); err != nil {
		return err
	}

	// Add a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	wg.Add(3)

	// Wait for the command to finish
	go func() {
		defer wg.Done()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		cmd.Wait()
	}()

	// Stdout handler
	go func() {
		defer wg.Done()
		defer stdoutPipe.Close()
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				output := make([]byte, n)
				copy(output, buf[:n])
				app.QueueUpdateDraw(func() {
					outputView.Write(output)
					outputView.ScrollToEnd()
				})
			}
			if err != nil {
				if err != io.EOF {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(outputView, "\nError: %v\n", err)
					})
				}
				break
			}
		}
	}()

	// Stderr handler
	go func() {
		defer wg.Done()
		defer stderrPipe.Close()
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				output := make([]byte, n)
				copy(output, buf[:n])
				app.QueueUpdateDraw(func() {
					outputView.Write(output)
					outputView.ScrollToEnd()
				})
			}
			if err != nil {
				if err != io.EOF {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(outputView, "\nError: %v\n", err)
					})
				}
				break
			}
		}
	}()

	wg.Wait()
	return nil
}
