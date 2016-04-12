package katool

//---------------------------------------------------------------------------
// Package factored import statement:
//---------------------------------------------------------------------------

import (

	// Stdlib
	"os"
	"os/exec"
)

//-----------------------------------------------------------------------------
// func: ExecutePipeline
//-----------------------------------------------------------------------------

func ExecutePipeline(cmd1, cmd2 *exec.Cmd) error {

	var err error

	// Adjust the stderr:
	cmd1.Stderr = os.Stderr
	cmd2.Stderr = os.Stderr

	// Connect both commands:
	cmd2.Stdin, err = cmd1.StdoutPipe()
	if err != nil {
		return err
	}

	// Execute the pipeline:
	if err := cmd2.Start(); err != nil {
		return err
	}
	if err := cmd1.Run(); err != nil {
		return err
	}
	if err := cmd2.Wait(); err != nil {
		return err
	}

	// Return on success:
	return nil
}
