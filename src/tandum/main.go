package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	running = true
)

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Println("Starting: ", arg)
		go keeprunning(arg)
	}
	select {}

}
func logfilename(arg string) string {
	hash := md5.Sum([]byte(arg))
	yymmddhhmm := time.Now().Format("2006010215")
	return "/tmp/tandrum_" + hex.EncodeToString(hash[:8]) + "_" + yymmddhhmm + ".log"
}

func keeprunning(arg string) {

	// Command to execute
	retries := 0
	started := time.Now()
	backoff := 1
	for running {
		logfile, err := os.OpenFile(logfilename(arg), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error opening log file:", err)
			os.Exit(1)
		}
		defer logfile.Close()

		retries += 1
		cmd := exec.Command("sh", "-c", arg)
		cmd.Stderr = logfile
		cmd.Stdout = logfile

		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting command:", err)
			os.Exit(1)
		}

		err = cmd.Wait()
		if err != nil {
			fmt.Println("Error:", err)

			if time.Since(started) < 10*time.Second && retries < 3 {
				fmt.Fprintln(os.Stderr, "Command failed immediately, not retrying")
				os.Exit(cmd.ProcessState.ExitCode())
			}

			fmt.Println("Retrying in", backoff, "seconds")
			time.Sleep(time.Duration(backoff) * time.Second)
			backoff = min(60, backoff*2)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
