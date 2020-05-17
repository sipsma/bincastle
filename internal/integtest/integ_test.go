package integtest

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/require"
)

type cmdOpt interface {
	applyTo(*exec.Cmd) error
}

type cmdOptFunc func(*exec.Cmd) error

func (f cmdOptFunc) applyTo(cmd *exec.Cmd) error {
	return f(cmd)
}

func bincastleArgs(gitUrl, gitRef, cmdPath string) cmdOpt {
	return cmdOptFunc(func(cmd *exec.Cmd) error {
		cmd.Args = append(cmd.Args[:1], "run", gitUrl, gitRef, cmdPath)
		return nil
	})
}

type cmdPty struct {
	parent *os.File
	child  *os.File
}

func (p cmdPty) Close() error {
	return multierror.Append(
		p.parent.Close(),
		p.child.Close(),
	).ErrorOrNil()
}

func (p cmdPty) applyTo(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.Stdin = p.child
	cmd.Stdout = p.child
	cmd.Stderr = p.child
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true
	cmd.SysProcAttr.Ctty = 0 // use the cmd's stdin fd as the ctty
	return nil
}

func getpty() (p cmdPty, err error) {
	parent, child, err := pty.Open()
	if err != nil {
		return p, err
	}
	return cmdPty{parent: parent, child: child}, nil
}

func withHomeDir(path string) cmdOpt {
	return cmdOptFunc(func(cmd *exec.Cmd) error {
		cmd.Dir = path
		cmd.Env = append(cmd.Env, "HOME="+path)
		return nil
	})
}

func withDebugStderr(dest io.Writer) cmdOpt {
	return cmdOptFunc(func(cmd *exec.Cmd) error {
		cmd.Stderr = nil
		stderrReader, err := cmd.StderrPipe()
		if err != nil {
			return err
		}
		go io.Copy(dest, stderrReader)

		return nil
	})
}

type bincastleCmd struct {
	*exec.Cmd
	startOnce sync.Once
	startErr  error
	waitOnce  sync.Once
	waitCh    chan struct{}
	waitErr   error
}

func (c *bincastleCmd) Start() error {
	c.startOnce.Do(func() {
		c.startErr = c.Cmd.Start()
	})
	return c.startErr
}

func (c *bincastleCmd) Wait(ctx context.Context) error {
	c.waitOnce.Do(func() {
		c.waitCh = make(chan struct{})
		go func() {
			defer close(c.waitCh)
			c.waitErr = c.Cmd.Wait()
		}()
	})

	select {
	case <-c.waitCh:
		return c.waitErr
	case <-ctx.Done():
		return ctx.Err()
	}
}

func newBincastleCmd(cmdCtx context.Context, cmdOpts ...cmdOpt) (*bincastleCmd, error) {
	p, err := exec.LookPath("bincastle")
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(cmdCtx, p)
	for _, o := range cmdOpts {
		err := o.applyTo(cmd)
		if err != nil {
			return nil, err
		}
	}

	// TODO bincastle shouldn't require this to be set in order to run
	cmd.Env = append(cmd.Env, "SSH_AUTH_SOCK="+os.Getenv("SSH_AUTH_SOCK"))

	return &bincastleCmd{Cmd: cmd}, nil
}

func TestDirectoriesRemoved(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO root dir should be configurable
	homeDir, err := ioutil.TempDir(filepath.Join(os.Getenv("HOME"), ".bincastle/test"), "")
	require.NoError(t, err)
	defer os.RemoveAll(homeDir)

	pty, err := getpty()
	require.NoError(t, err)
	defer pty.Close()

	// TODO don't rely on remote resources during test runtime
	bcCmd, err := newBincastleCmd(ctx,
		bincastleArgs(
			"https://github.com/sipsma/bincastle.git", "3da9a59", "internal/integtest/stubsystem"),
		withHomeDir(homeDir),
		pty,
		withDebugStderr(os.Stdout),
	)
	require.NoError(t, err)

	scanReader, scanWriter := io.Pipe()
	go func() {
		defer scanWriter.Close()
		io.Copy(io.MultiWriter(os.Stdout, scanWriter), pty.parent)
	}()

	outputScanner := bufio.NewScanner(scanReader)
	outputScanner.Split(bufio.ScanWords)

	scanCh := make(chan bool)
	go func() {
		for outputScanner.Scan() {
			if outputScanner.Text() == "BINCASTLEINITIALIZED" {
				scanCh <- true
				break
			}
		}
		scanCh <- false
	}()

	err = bcCmd.Start()
	require.NoError(t, err)

	waitCh := make(chan error)
	go func() {
		defer close(waitCh)
		waitCh <- bcCmd.Wait(ctx)
	}()

	select {
	case waitErr := <-waitCh:
		require.FailNow(t, "bincastle unexpectedly exited", "exit err: %v", waitErr)
	case foundSentinel := <-scanCh:
		require.True(t, foundSentinel)
	}

	innerRuncRootDir := filepath.Join(homeDir,
		".bincastle/var/lib/buildkitd/runc-overlayfs/execs/testctr/testctr")
	_, err = os.Stat(innerRuncRootDir)
	require.NoError(t, err)

	outerRuncRootDir := filepath.Join(homeDir, ".bincastle/ctrs/system/system")
	_, err = os.Stat(outerRuncRootDir)
	require.NoError(t, err)

	err = pty.parent.Close()
	require.NoError(t, err)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 15*time.Second)
	defer timeoutCancel()
	err = bcCmd.Wait(timeoutCtx)
	require.NotEqual(t, context.DeadlineExceeded, err)
	require.NotEqual(t, context.Canceled, err)

	_, err = os.Stat(innerRuncRootDir)
	require.True(t, os.IsNotExist(err), "unexpected err: %v", err)

	_, err = os.Stat(outerRuncRootDir)
	require.True(t, os.IsNotExist(err), "unexpected err: %v", err)
}
