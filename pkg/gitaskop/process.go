package gitaskop

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

var (
	hackLockIsRunning = false
)

func getJob(task Task) (Job, error) {
	folder := getRepoFolder(task)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return Job{}, err
	}
	
	file, err := os.Open(folder + "/.gitaskop.json")
	if err != nil {
		return Job{}, err
	}
	defer file.Close()
	
	var job Job
	if err := json.NewDecoder(file).Decode(&job); err != nil {
		return Job{}, err
	}
	return job, nil
}

func tryReload(task *Task, job *Job) error {
	if err := pullRepo(*task); err != nil {
		fmt.Println("Error in pulling repo:", task.RepoURL, err)
		return err
	}

	newjob, err := getJob(*task)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Println(err)
		return err
	}
	
	if newjob.Interval != job.Interval || newjob.Script != job.Script {
		log.Println("Job updated")
		*job = newjob
		crond.Stop()
		crond.Remove(cron.EntryID(0))
		crond.AddFunc(job.Interval, jobFunc)
		crond.Start()
	}
	return nil
}

func createCronFunc(task *Task, job *Job, oneshot bool, triggerOnChange bool) func() {
	return func() {
		if hackLockIsRunning {
			fmt.Println("Job already running")
			return
		}
		hackLockIsRunning = true
		defer func() {
			hackLockIsRunning = false
			fmt.Println("hacklock released")
		}()
		
		runid := uuid.New().String()
		fmt.Println("Running job", task.Name, "with id", runid)
		defer fmt.Println("Finish job", task.Name, "with id", runid)

		rev := getRevision(*task)
		if triggerOnChange && rev != "" {
			if rev == getSavedRevision(*task) {
				fmt.Println("No change in repo")
				if !oneshot {
					tryReload(task, job)
				}
				return
			}
		}
		saveRevision(*task)

		cmd := exec.Command("sh", "-c", job.Script)
		cmd.Dir = getRepoFolder(*task)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "cron failed for ", task.RepoURL, err)
		}
		
		if !oneshot {
			tryReload(task, job)
		}
	}
}