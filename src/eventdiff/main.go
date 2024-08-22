package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

var (
	lines                  map[string]string = make(map[string]string)
	output_on_first_change bool              = true
	use_file_cache         bool              = true
	cache_path             string            = "/tmp/eventdiff_cache"
	remove_cache_onstart   bool              = true
)

func exists(key string) (bool, error) {
	if use_file_cache {
		_, err := os.Stat(cache_path + "/" + key + ".eventdiff")
		//fixme: report non missing file errors
		return err == nil, nil
	}

	_, exists := lines[key]
	return exists, nil
}

func set(key string, value string) error {
	if use_file_cache {
		return os.WriteFile(cache_path+"/"+key+".eventdiff", []byte(value), 0644)
	}
	lines[key] = value
	return nil
}

func ifDiff(key string, value string) (bool, error) {

	exists, err := exists(key)
	if err != nil {
		return false, err
	}
	if !exists {
		set(key, value)
		return output_on_first_change, nil
	}
	if use_file_cache {
		content, err := os.ReadFile(cache_path + "/" + key + ".eventdiff")
		if err != nil {
			return false, err
		}
		if string(content) != value {
			set(key, value)
			return true, nil
		}
		return false, nil
	}

	if lines[key] != value {
		set(key, value)
		return true, nil
	}
	return false, nil
}

func main() {

	// setup
	if use_file_cache {
		if _, err := os.Stat(cache_path); os.IsNotExist(err) {
			os.Mkdir(cache_path, 0755)
		}
		if remove_cache_onstart {
			files, _ := os.ReadDir(cache_path)
			for _, file := range files {
				if strings.HasSuffix(file.Name(), ".eventdiff") {
					os.Remove(cache_path + "/" + file.Name())
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

		//split line on first instance of space
		key := strings.Split(line, "|")[0]
		value := line[len(key)+1:]

		diff, err := ifDiff(key, value)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
		}

		if diff {
			fmt.Println(line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
