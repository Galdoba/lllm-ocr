package ocr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/Galdoba/lllm-ocr/ocr/cycle"
)

// RecognizePages runs llama-mtmd-cli on each PNG in pages, producing per-page HTML files.
func RecognizePages(cfg Config, pages []int) error {
	// Go 1.22+ per-iteration loop vars: `p` is safely captured per iteration.
	for _, p := range pages {
		if cfg.ProgressFn != nil {
			cfg.ProgressFn(p, len(pages))
		}

		if err := RecognizeSinglePage(cfg, p); err != nil {
			return fmt.Errorf("recognizing page %d: %w", p, err)
		}
	}

	return nil
}

// RecognizeSinglePage processes a single page with retry support.
func RecognizeSinglePage(cfg Config, pageNum int) error {
	pn := pageName(pageNum)
	imgPath := fmt.Sprintf("%s/%s.png", cfg.OutDir, pn)
	htmlPath := fmt.Sprintf("%s/%s.html", cfg.OutDir, pn)

	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		return fmt.Errorf("image file %s not found for page %d", imgPath, pageNum)
	}

	// Initialize retry state
	maxRetries := cfg.Recognize.MaxRetries
	if maxRetries < 0 {
		maxRetries = 0
	}
	retryState := NewRetryState(maxRetries, cfg.Recognize.LlamaCliArgs, cfg.Recognize.RetryTempArgs)

	for retryState.CanRetry() {
		// Create HTML file for this attempt
		htmlFile, err := os.Create(htmlPath)
		if err != nil {
			return fmt.Errorf("creating HTML file for page %d: %w", pageNum, err)
		}

		// Get current args (original on first attempt, adjusted on retry)
		args := getArgsForAttempt(retryState)
		args = append(args, "--image", imgPath)

		// Create command with process group for clean killing
		cmd := exec.Command(cfg.Recognize.LlamaCliCmd, args...)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		// Setup stdout pipe
		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			htmlFile.Close()
			return fmt.Errorf("creating stdout pipe for page %d: %w", pageNum, err)
		}

		// Initialize fresh detector for each attempt
		detector := cycle.NewDetector(cfg.Cycle)

		// Setup output writer
		out := io.MultiWriter(htmlFile, os.Stderr)

		// Channel for cycle detection signal
		cycleDetected := make(chan bool, 1)

		// WaitGroup to synchronize goroutine completion
		var wg sync.WaitGroup
		wg.Add(1)

		// Start command
		if err := cmd.Start(); err != nil {
			htmlFile.Close()
			return fmt.Errorf("starting llama-mtmd-cli for page %d: %w", pageNum, err)
		}

		// Goroutine ONLY reads tokens and detects cycles
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Page %d: stdout reader panic: %v", pageNum, r)
					select {
					case cycleDetected <- false:
					default:
					}
				}
			}()

			scanner := bufio.NewScanner(stdoutPipe)
			scanner.Split(bufio.ScanWords)

			for scanner.Scan() {
				token := scanner.Text()
				detector.AddToken(token)
				fmt.Fprint(out, token, " ")

				if detector.IsCycling() {
					// Kill entire process group to close all inherited pipes
					if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
						log.Printf("Page %d: warning - failed to kill process group: %v", pageNum, err)
					}
					select {
					case cycleDetected <- true:
					default:
					}
					return
				}
			}
			select {
			case cycleDetected <- false:
			default:
			}
		}()

		// Main goroutine waits for command to finish
		cmdErr := cmd.Wait()

		// Wait for reader goroutine to finish (ensures all writes complete)
		wg.Wait()

		// Blocking receive for cycle signal (never lost)
		isCycling := false
		select {
		case isCycling = <-cycleDetected:
		default:
		}

		// Close HTML file after goroutine finishes
		htmlFile.Close()

		if !isCycling {
			// No cycle detected - success or normal error
			if cmdErr != nil {
				if !isProcessKilledError(cmdErr) {
					log.Printf("Page %d: warning - command error: %v", pageNum, cmdErr)
				}
			}
			// Complete HTML if there was an error (crash recovery)
			if cmdErr != nil {
				if err := cycle.CompleteHTMLWithContent(htmlPath, true); err != nil {
					log.Printf("Page %d: warning - failed to complete HTML: %v", pageNum, err)
				}
			}
			return nil
		}

		// Cycle was detected - prepare for retry
		log.Printf("Page %d: cycle detected, retrying...", pageNum)
		retryState.IncrementRetry()
	}

	// Max retries exceeded - complete with available content
	log.Printf("Page %d: max retries (%d) exceeded, completing with available content", pageNum, retryState.MaxRetries)
	if err := cycle.CompleteHTMLWithContent(htmlPath, true); err != nil {
		log.Printf("Page %d: warning - failed to complete HTML: %v", pageNum, err)
	}
	return nil
}

// getArgsForAttempt returns the CLI args for the current retry attempt.
func getArgsForAttempt(s *RetryState) []string {
	// On first attempt (CurrentRetry == 0), use original args
	if s.CurrentRetry == 0 {
		result := make([]string, len(s.OriginalArgs))
		copy(result, s.OriginalArgs)
		return result
	}

	// On retry attempts, apply temperature adjustment
	newArgs := make([]string, len(s.OriginalArgs))
	copy(newArgs, s.OriginalArgs)

	foundTemp := false
	for i, arg := range newArgs {
		if arg == "--temp" && i+1 < len(newArgs) {
			if len(s.RetryTempArgs) > 1 {
				newArgs[i+1] = s.RetryTempArgs[1]
			}
			foundTemp = true
			break
		}
	}
	if !foundTemp && len(s.RetryTempArgs) >= 2 {
		newArgs = append(newArgs, s.RetryTempArgs...)
	}

	return newArgs
}

// isProcessKilledError returns true if the error indicates the process was killed
// (exit code 137 = SIGKILL, or os.ErrProcessDone).
func isProcessKilledError(err error) bool {
	if errors.Is(err, os.ErrProcessDone) {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			return status.Signal() == syscall.SIGKILL || status.ExitStatus() == 137
		}
	}
	return false
}
