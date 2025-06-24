package gitaskop

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/robfig/cron/v3"
)

// Task represents a git task configuration
type Task struct {
	Name    string
	RepoURL string
	Branch  string
}

// Job represents a cron job configuration
type Job struct {
	Interval string
	Script   string
}

var (
	jobFunc = func() {}
	crond   = cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
)

// Run is the main entry point for gitaskop functionality
func Run(args []string) error {
	// Set up default values
	job := Job{
		Interval: "*/5 * * * *",
		Script:   "echo 'hello'",
	}
	task := Task{
		Name:    "task",
		RepoURL: "file:///tmp/gr",
		Branch:  "main",
	}

	// Set up flag sets
	jobCmd := flag.NewFlagSet("job", flag.ExitOnError)
	taskCmd := flag.NewFlagSet("task", flag.ExitOnError)
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	cleanCmd := flag.NewFlagSet("clean", flag.ExitOnError)

	subcommands := map[string]*flag.FlagSet{
		jobCmd.Name():   jobCmd,
		taskCmd.Name():  taskCmd,
		runCmd.Name():   runCmd,
		cleanCmd.Name(): cleanCmd,
	}

	// Parse arguments
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: gitaskop <subcommand> [options]")
		fmt.Fprintln(os.Stderr, "Subcommands:")
		fmt.Fprintln(os.Stderr, "  run")
		fmt.Fprintln(os.Stderr, "  clean")
		fmt.Fprintln(os.Stderr, "  generate")
		return fmt.Errorf("insufficient arguments")
	}

	maincommand := args[0]
	command := maincommand
	
	if maincommand == "generate" && len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gitaskop generate <task|job>")
		return fmt.Errorf("generate requires task or job argument")
	}
	
	if (maincommand == "generate" || maincommand == "run") && len(args) > 1 {
		if args[1] == "task" {
			command = "task"
		} else if args[1] == "job" {
			command = "job"
		}
	}

	// Set up flags for each command
	for _, fs := range subcommands {
		if maincommand == "run" {
			fs.Bool("oneshot", false, "Run once and exit")
			fs.Bool("clean", false, "Clean the tmp folder")
			fs.Bool("alwaysclone", false, "Always clone the repo")
			fs.Bool("triggeronchange", false, "only trigger job if there are changes in the repo")
		}
		if command == "job" || maincommand == "run" {
			fs.String("interval", job.Interval, "cron interval")
			fs.String("script", job.Script, "command to run")
		}
		if command == "task" || maincommand == "run" {
			fs.String("name", task.Name, "task name")
			fs.String("repo", task.RepoURL, "repo url")
			fs.String("branch", task.Branch, "branch")
		}
	}

	cmd := subcommands[command]
	if cmd == nil {
		return fmt.Errorf("unknown subcommand '%s', see help for more details", args[0])
	}

	// Parse remaining arguments
	var parseArgs []string
	if maincommand != command {
		parseArgs = args[2:]
	} else {
		parseArgs = args[1:]
	}
	
	if err := cmd.Parse(parseArgs); err != nil {
		return err
	}

	// Update task and job from flags
	if command == "task" || maincommand == "run" {
		task = Task{
			Name:    cmd.Lookup("name").Value.String(),
			RepoURL: cmd.Lookup("repo").Value.String(),
			Branch:  cmd.Lookup("branch").Value.String(),
		}
		if maincommand == "generate" {
			return generateTask(&task)
		}
	} else if command == "job" || maincommand == "run" {
		job = Job{
			Interval: cmd.Lookup("interval").Value.String(),
			Script:   cmd.Lookup("script").Value.String(),
		}
		if maincommand == "generate" {
			return generateJob(&job)
		}
	}

	// Check flags
	flagset := make(map[string]bool)
	cmd.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	oneshot := flagset["oneshot"]
	clean := flagset["clean"] || command == "clean"
	alwaysclone := flagset["alwaysclone"]
	triggeronchange := flagset["triggeronchange"]

	// Set up signal handling
	c := make(chan os.Signal, 1)
	if runtime.GOOS == "linux" {
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	} else {
		signal.Notify(c, os.Interrupt)
	}

	go catchSignals(c, task, crond, clean)

	if clean {
		if err := cleanRepo(task); err != nil {
			fmt.Fprintln(os.Stderr, "Failed to clean repo:", err)
		}
		defer cleanRepo(task)
	}

	if command == "run" {
		if err := cloneRepo(task, alwaysclone); err != nil {
			fmt.Fprintln(os.Stderr, "Error in cloning repo")
			return err
		}
		
		jobConfig, err := getJob(task)
		fmt.Println("job:", jobConfig)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error in getting job")
			return err
		}

		jobFunc = createCronFunc(&task, &jobConfig, oneshot, triggeronchange)
		if oneshot {
			jobFunc()
			return nil
		}
		
		crond.AddFunc(jobConfig.Interval, jobFunc)
		crond.Start()
		// Keep running until signal
		select {}
	}

	return nil
}

func catchSignals(c chan os.Signal, task Task, crond *cron.Cron, clean bool) {
	for sig := range c {
		fmt.Println("Received signal:", sig)
		crond.Stop()
		if clean {
			if err := cleanRepo(task); err != nil {
				os.Exit(1)
			}
		}
		os.Exit(0)
	}
}

func generateTask(task *Task) error {
	str, err := json.Marshal(task)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fmt.Fprintln(os.Stdout, string(str))
	return nil
}

func generateJob(job *Job) error {
	str, err := json.Marshal(job)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fmt.Fprintln(os.Stdout, string(str))
	return nil
}