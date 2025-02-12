package services

import (
	"bbrew/internal/models"
	"github.com/rivo/tview"
	"io"
	"os/exec"
)

type CommandServiceInterface interface {
	UpdateAllPackages(app *tview.Application, outputView *tview.TextView) error
	UpdatePackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error
	RemovePackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error
	InstallPackage(info models.Formula, app *tview.Application, outputView *tview.TextView) error
	UpdateHomebrew(app *tview.Application, outputView *tview.TextView) error
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

func (s *CommandService) UpdateHomebrew(app *tview.Application, outputView *tview.TextView) error {
	cmd := exec.Command("brew", "update")
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

	go func() {
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		cmd.Wait()
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stdoutPipe.Read(buf)
			if n > 0 {
				app.QueueUpdateDraw(func() {
					outputView.Write(buf[:n])
					outputView.ScrollToEnd()
				})
			}
			if err != nil {
				break
			}
		}
	}()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := stderrPipe.Read(buf)
			if n > 0 {
				app.QueueUpdateDraw(func() {
					outputView.Write(buf[:n])
					outputView.ScrollToEnd()
				})
			}
			if err != nil {
				break
			}
		}
	}()

	cmd.Wait()

	return nil
}
