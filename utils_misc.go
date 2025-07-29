package gobox_utils

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// calcPercent calculates the percentage progress based on current index and total count.
// Parameters:
//   - i: current zero-based index (e.g., 0 for first item).
//   - total: total number of items.
//
// Returns:
//   - int percentage value representing progress from 1 to 100.
//     If total is zero, returns 100 to avoid division by zero.
func calcPercent(i, total int) int {
	if total == 0 {
		// Avoid division by zero; assume 100% when total is zero
		return 100

	}
	// Calculate progress percentage: (current index + 1) / total * 100
	return int(float64(i+1) / float64(total) * 100)
}

type TerminalSession struct {
	cmd      *exec.Cmd
	ptyFile  *os.File
	writeMu  sync.Mutex
	ctx      context.Context
	IsActive bool
}

var terminalInstance *TerminalSession

func StartTerminalStream(containerName string, ctx context.Context) error {
	if terminalInstance != nil && terminalInstance.IsActive {
		return errors.New("terminal session already running")
	}

	println("Starting terminal stream...")
	cmd := exec.Command("docker", "exec", "-it", containerName, "bash")

	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	terminalInstance = &TerminalSession{
		cmd:      cmd,
		ptyFile:  ptyFile,
		ctx:      ctx,
		IsActive: true,
	}

	// Start reading from terminal and emit data to frontend
	go func() {
		scanner := bufio.NewScanner(ptyFile)
		scanner.Split(bufio.ScanBytes)

		for scanner.Scan() {
			runtime.EventsEmit(ctx, "terminal:data", scanner.Text())
		}

		terminalInstance.IsActive = false
	}()

	//  Listen for user input and write it to the PTY
	runtime.EventsOn(ctx, "terminal:input", func(data ...interface{}) {
		if len(data) == 0 {
			return
		}
		input, ok := data[0].(string)
		if ok {
			WriteToTerminal(input)
		}
	})

	return nil
}

func WriteToTerminal(input string) error {
	if terminalInstance == nil || !terminalInstance.IsActive {
		return errors.New("no terminal session running")
	}

	terminalInstance.writeMu.Lock()
	defer terminalInstance.writeMu.Unlock()

	_, err := terminalInstance.ptyFile.Write([]byte(input))
	return err
}
