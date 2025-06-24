package tandum

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Config holds configuration for tandum
type Config struct {
	Running bool
}

// Run starts tandum with the given commands
func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no commands provided")
	}

	config := &Config{Running: true}
	
	// Set up signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	var wg sync.WaitGroup
	
	go func() {
		<-c
		fmt.Println("\nReceived interrupt signal, shutting down...")
		config.Running = false
		wg.Wait()
		os.Exit(0)
	}()

	// Start each command in a goroutine
	for _, arg := range args {
		fmt.Println("Starting:", arg)
		wg.Add(1)
		go func(command string) {
			defer wg.Done()
			keepRunning(command, config)
		}(arg)
	}

	// Wait indefinitely
	select {}
}

func logFilename(arg string) string {
	hash := md5.Sum([]byte(arg))
	yymmddhhmm := time.Now().Format("2006010215")
	return "/tmp/tandrum_" + hex.EncodeToString(hash[:8]) + "_" + yymmddhhmm + ".log"
}

func keepRunning(arg string, config *Config) {
	retries := 0
	started := time.Now()
	backoff := 1
	
	for config.Running {
		logfile, err := os.OpenFile(logFilename(arg), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
			return
		}

		retries++
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = logfile
		cmd.Stdout = logfile

		err = cmd.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting command: %v\n", err)
			logfile.Close()
			return
		}

		err = cmd.Wait()
		logfile.Close()
		
		if err != nil {
			fmt.Printf("Command failed: %v\n", err)

			if time.Since(started) < 10*time.Second && retries < 3 {
				fmt.Fprintln(os.Stderr, "Command failed immediately, not retrying")
				return
			}

			if !config.Running {
				return
			}

			fmt.Printf("Retrying in %d seconds\n", backoff)
			time.Sleep(time.Duration(backoff) * time.Second)
			backoff = min(60, backoff*2)
		} else {
			// Command completed successfully, reset backoff
			backoff = 1
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}