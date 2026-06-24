package services

import (
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/rivo/tview"
)

// ExecuteCommand runs a command, streaming stdout/stderr to a tview.TextView in real time.
func ExecuteCommand(app *tview.Application, cmd *exec.Cmd, outputView *tview.TextView) error {
	stdoutPipe, stdoutWriter := io.Pipe()
	stderrPipe, stderrWriter := io.Pipe()
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(3)

	cmdErrCh := make(chan error, 1)

	go func() {
		defer wg.Done()
		defer stdoutWriter.Close()
		defer stderrWriter.Close()
		cmdErrCh <- cmd.Wait()
	}()

	streamPipe := func(pipe *io.PipeReader) {
		defer wg.Done()
		defer pipe.Close()
		buf := make([]byte, 1024)
		for {
			n, err := pipe.Read(buf)
			if n > 0 {
				output := make([]byte, n)
				copy(output, buf[:n])
				app.QueueUpdateDraw(func() {
					_, _ = outputView.Write(output) // #nosec G104
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
	}

	go streamPipe(stdoutPipe)
	go streamPipe(stderrPipe)

	wg.Wait()

	return <-cmdErrCh
}
