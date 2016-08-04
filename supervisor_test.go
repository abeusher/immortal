package immortal

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

type myFork struct{}

func (self myFork) Fork() {
	return
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGUSR1)
	select {
	case s := <-c:
		if s != syscall.SIGHUP {
			return
		}
	}
}

func TestSupervisor(t *testing.T) {
	base := filepath.Base(os.Args[0]) // "exec.test"
	dir := filepath.Dir(os.Args[0])   // "/tmp/go-buildNNNN/os/exec/_test"
	if dir == "." {
		t.Skip("skipping; running test at root somehow")
	}
	parentDir := filepath.Dir(dir) // "/tmp/go-buildNNNN/os/exec"
	dirBase := filepath.Base(dir)  // "_test"
	if dirBase == "." {
		t.Skipf("skipping; unexpected shallow dir of %q", dir)
	}

	cfg := &Config{
		Env:     map[string]string{"GO_WANT_HELPER_PROCESS": "1"},
		command: []string{filepath.Join(dirBase, base), "-test.run=TestHelperProcess"},
		Cwd:     parentDir,
	}
	d := &Daemon{
		Config: cfg,
		Control: &Control{
			fifo:  make(chan Return),
			quit:  make(chan struct{}),
			state: make(chan error),
		},
		Forker: &myFork{},
		Logger: &LogWriter{
			logger: NewLogger(cfg),
		},
	}
	d.Run()
	select {
	case err := <-d.Control.state:
		if err == nil {
			t.Error("Expecting error: signal: Killed")
			return
		}
	case <-time.After(1 * time.Second):
		d.process.Kill()
	}
}
