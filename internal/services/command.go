package services

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// ExecuteCommand runs a command, streaming stdout/stderr to the provided writer in real time.
func ExecuteCommand(cmd *exec.Cmd, output io.Writer) error {
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
				_, _ = output.Write(buf[:n]) // #nosec G104
			}
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(output, "\nError: %v\n", err)
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
