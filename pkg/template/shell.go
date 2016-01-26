package template

import (
	"bytes"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"io"
	"os"
	"os/exec"
)

func ExecuteShell(ctx context.Context) interface{} {
	// TODO - copy stdout to buffer and return as string
	return func(line string) (io.Reader, error) {
		c := exec.Command("sh", "-")

		output := new(bytes.Buffer)
		if stdout, err := c.StdoutPipe(); err == nil {
			fanout := io.MultiWriter(os.Stdout, output)
			go func() {
				io.Copy(fanout, stdout)
			}()
		} else {
			return nil, err
		}

		if stderr, err := c.StderrPipe(); err == nil {
			go func() {
				io.Copy(os.Stderr, stderr)
			}()
		} else {
			return nil, err
		}
		stdin, err := c.StdinPipe()
		if err != nil {
			return nil, err
		}
		if err := c.Start(); err != nil {
			return nil, err
		}
		if _, err := stdin.Write([]byte(line)); err != nil {
			stdin.Close()
			return nil, err
		}
		stdin.Close() // finished
		err = c.Wait()
		if ee, ok := err.(*exec.ExitError); ok {
			glog.Infoln("PID", ee.Pid(), " - Process state", ee.Success())
		}
		return output, err
	}
}
