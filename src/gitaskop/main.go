package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/robfig/cron/v3"
)

// TODO: pass these as env vars
type Task struct {
	Name    string
	RepoURL string
	Branch  string
}

// TODO: pass these as env vars
type Job struct {
	Interval string
	Script   string
}

var (
	jobFunc = func() {}
	crond   = cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	job = Job{
		Interval: "*/5 * * * *",
		Script:   "echo 'hello'",
	}
	task = Task{
		Name:    "task",
		RepoURL: "file:///tmp/gr",
		Branch:  "main",
	}
	jobCmd   = flag.NewFlagSet("job", flag.ExitOnError)
	taskCmd  = flag.NewFlagSet("task", flag.ExitOnError)
	runCmd   = flag.NewFlagSet("run", flag.ExitOnError)
	cleanCmd = flag.NewFlagSet("clean", flag.ExitOnError)

	subcommands = map[string]*flag.FlagSet{
		jobCmd.Name():   jobCmd,
		taskCmd.Name():  taskCmd,
		runCmd.Name():   runCmd,
		cleanCmd.Name(): cleanCmd,
	}
	triggeronchange bool = false
)

func main() {

	//TODO: run with no args, with sane defaults (if files exists)
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: gitaskop <subcommand> [options]")
		fmt.Fprintln(os.Stderr, "Subcommands:")
		fmt.Fprintln(os.Stderr, "  run")
		fmt.Fprintln(os.Stderr, "  clean")
		fmt.Fprintln(os.Stderr, "  generate")
		os.Exit(1)
	}
	maincommand := os.Args[1]
	command := maincommand
	if maincommand == "generate" && len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: gitaskop generate <task|job>")
		os.Exit(1)
	}
	if (maincommand == "generate" || maincommand == "run") && len(os.Args) > 2 {
		if os.Args[2] == "task" {
			command = "task"
		} else if os.Args[2] == "job" {
			command = "job"
		}
	}

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
		log.Fatalf("[ERROR] unknown subcommand '%s', see help for more details.", os.Args[1])
	}
	// fmt.Println("maincommand:", maincommand)
	// fmt.Println("command:", command)
	if maincommand != command {
		cmd.Parse(os.Args[3:])
	} else {
		cmd.Parse(os.Args[2:])
	}
	//command
	// command := cmd.Name()
	// fmt.Println(command)

	// generate := false
	// run := false //thing about --dry-run

	//TODO: validate inputs??
	if command == "task" || maincommand == "run" {
		task = Task{
			Name:    cmd.Lookup("name").Value.String(),
			RepoURL: cmd.Lookup("repo").Value.String(),
			Branch:  cmd.Lookup("branch").Value.String(),
		}
		if maincommand == "generate" {
			generateTask(&task)
		}
	} else if command == "job" || maincommand == "run" {
		job = Job{
			//TODO: validate inputs
			Interval: cmd.Lookup("interval").Value.String(),
			Script:   cmd.Lookup("script").Value.String(),
		}
		if maincommand == "generate" {
			generateJob(&job)
		}
	}
	// else
	// fmt.Fprintln(os.Stderr, "generate task or job")
	// os.Exit(1)
	flagset := make(map[string]bool)
	cmd.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	oneshot := flagset["oneshot"]
	clean := flagset["clean"] || command == "clean"
	alwaysclone := flagset["alwaysclone"] //BUG: alwaysclone is not working
	triggeronchange = flagset["triggeronchange"]

	// catch ctrl+c and kill signals
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
			log.Panic(err)
		}
		job, err := getJob(task)
		fmt.Println("job:", job)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error in getting job")
			log.Panic(err)
		}

		jobFunc = createCronFunc(&task, &job, oneshot)
		if oneshot {
			jobFunc()
			return
		}
		crond.AddFunc(job.Interval, jobFunc)
		crond.Start()
		// goroutine (cron) running
		select {}
	}
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

func generateTask(task *Task) {
	if str, err := json.Marshal(&task); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintln(os.Stdout, string(str))
	}
	os.Exit(0)

}
func generateJob(job *Job) {
	if str, err := json.Marshal(&job); err != nil {
		fmt.Fprintln(os.Stderr, err)
	} else {
		fmt.Fprintln(os.Stdout, string(str))
	}
	os.Exit(0)
}
