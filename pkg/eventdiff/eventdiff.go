package eventdiff

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds configuration for eventdiff
type Config struct {
	OutputOnFirstChange bool
	UseFileCache       bool
	CachePath          string
	RemoveCacheOnStart bool
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		OutputOnFirstChange: true,
		UseFileCache:       true,
		CachePath:          "/tmp/eventdiff_cache",
		RemoveCacheOnStart: true,
	}
}

// EventDiff holds the state for event difference detection
type EventDiff struct {
	config Config
	lines  map[string]string
}

// New creates a new EventDiff instance
func New(config Config) *EventDiff {
	return &EventDiff{
		config: config,
		lines:  make(map[string]string),
	}
}

// Run starts the eventdiff processing
func Run(args []string) error {
	config := DefaultConfig()
	
	// TODO: Parse args for configuration options
	
	ed := New(config)
	return ed.Process()
}

// Process handles the main event difference detection logic
func (ed *EventDiff) Process() error {
	// Setup
	if ed.config.UseFileCache {
		if _, err := os.Stat(ed.config.CachePath); os.IsNotExist(err) {
			os.Mkdir(ed.config.CachePath, 0755)
		}
		if ed.config.RemoveCacheOnStart {
			files, _ := os.ReadDir(ed.config.CachePath)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".eventdiff") {
					os.Remove(ed.config.CachePath + "/" + file.Name())
				}
			}
		}
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "|") {
			fmt.Fprintln(os.Stderr, "No pipe found")
			continue
		}

		// Split line on first instance of pipe
		key := strings.Split(line, "|")[0]
		value := line[len(key)+1:]

		diff, err := ed.ifDiff(key, value)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}

		if diff {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		return err
	}
	
	return nil
}

func (ed *EventDiff) exists(key string) (bool, error) {
	if ed.config.UseFileCache {
		_, err := os.Stat(ed.config.CachePath + "/" + key + ".eventdiff")
		return err == nil, nil
	}

	_, exists := ed.lines[key]
	return exists, nil
}

func (ed *EventDiff) set(key string, value string) error {
	if ed.config.UseFileCache {
		return os.WriteFile(ed.config.CachePath+"/"+key+".eventdiff", []byte(value), 0644)
	}
	ed.lines[key] = value
	return nil
}

func (ed *EventDiff) ifDiff(key string, value string) (bool, error) {
	exists, err := ed.exists(key)
	if err != nil {
		return false, err
	}
	if !exists {
		ed.set(key, value)
		return ed.config.OutputOnFirstChange, nil
	}
	if ed.config.UseFileCache {
		content, err := os.ReadFile(ed.config.CachePath + "/" + key + ".eventdiff")
		if err != nil {
			return false, err
		}
		if string(content) != value {
			ed.set(key, value)
			return true, nil
		}
		return false, nil
	}

	if ed.lines[key] != value {
		ed.set(key, value)
		return true, nil
	}
	return false, nil
}