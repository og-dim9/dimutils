package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-cmd/cmd"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Input          string `yaml:"input"`
	InputFile      string `yaml:"inputFile"`
	Output         string `yaml:"output"`
	OutputFile     string `yaml:"outputFile"`
	Command        string `yaml:"command"`
	Name           string `yaml:"name"`
	Stdout         string
	Stderr         string
	StartedAt      time.Time
	EndedAt        time.Time
	ExpectedOutput string
	Error          error
}

var (
	print_exit_code_only = os.Getenv("UNEXPECT_PRINTEXITCODEONLY") != ""
)

func main() {

	// Read the YAML file
	yamlFile, err := ioutil.ReadFile("unexpect.yaml")
	if err != nil {
		log.Fatal(err)
	}

	// Parse the YAML file into a Config struct
	var configs []Config
	err = yaml.Unmarshal(yamlFile, &configs)
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(len(configs))

	for i := 0; i < len(configs); i++ {
		config := &configs[i]
		go func() {
			defer wg.Done()

			config.StartedAt = time.Now()
			err := test(config)
			config.Error = err
			config.EndedAt = time.Now()

		}()
	}

	wg.Wait()

	results(configs)
	os.Exit(returnCode(configs))
}

func test(config *Config) error {

	config.StartedAt = time.Now()

	expected, err := getExpectedOutput(config)

	config.ExpectedOutput = expected

	if err != nil {
		return err
	}

	statusChan, err := getStartCommand(config)

	if err != nil {
		return err
	}

	for status := range statusChan {
		if status.Error != nil {
			return status.Error
		}

		if status.Complete {
			config.Stdout = strings.Join(status.Stdout, "")
			config.Stderr = strings.Join(status.Stderr, "")
			break
		}
	}

	return nil
}

func getExpectedOutput(config *Config) (string, error) {

	if config.OutputFile != "" && config.Output != "" {
		return "", fmt.Errorf("Both output and outputFile are defined")
	}

	if config.OutputFile != "" {
		output, err := ioutil.ReadFile(config.OutputFile)
		if err != nil {
			return string(output), err
		}
		return string(output), nil
	}

	return config.Output, nil
}

func getStartCommand(config *Config) (<-chan cmd.Status, error) {

	if config.Input != "" && config.InputFile != "" {
		return nil, fmt.Errorf("Both input and inputFile are defined")
	}
	if config.Command == "" {
		return nil, fmt.Errorf("Command is not defined")
	}

	testCmd := cmd.NewCmd("sh", "-c", config.Command)

	if config.InputFile != "" {
		input, err := ioutil.ReadFile(config.InputFile)
		if err != nil {
			return nil, err
		}
		return testCmd.StartWithStdin(strings.NewReader(string(input))), nil
	}

	return testCmd.StartWithStdin(strings.NewReader(config.Input)), nil
}

// FIXME: move to a separate file and add junit output
func results(configs []Config) {
	if print_exit_code_only {
		fmt.Println(returnCode(configs))
		return
	}
	for _, config := range configs {
		fmt.Println("-----------------------------")
		fmt.Println("Name:", config.Name)
		fmt.Println("Passed:", config.ExpectedOutput == config.Stdout)
		if config.ExpectedOutput != config.Stdout {
			//FIXME: print failure to stderr
			fmt.Println("Expected:-")
			fmt.Println(config.ExpectedOutput)
			fmt.Println("Got:-")
			fmt.Println(config.Stdout)
		}

		fmt.Println("Elapsed Time:", config.EndedAt.Sub(config.StartedAt))
		if config.Error != nil {
			fmt.Println("Error:", config.Error)
		}
		fmt.Println("Command:", config.Command)
	}
	fmt.Println("-----------------------------")
}

func returnCode(configs []Config) int {

	for _, config := range configs {
		if config.ExpectedOutput != config.Stdout {
			return 1
		}
	}
	return 0
}
