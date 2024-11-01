package main

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
	HACKLOCK_is_running = false
)

func getJob(task Task) (Job, error) {
	folder := getRepoFolder(task)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return Job{}, err
	}
	//open file in repo
	file, err := os.Open(folder + "/.gitaskop.json")
	if err != nil {
		return Job{}, err
	}
	defer file.Close()
	// Decode the file
	var job Job
	if err := json.NewDecoder(file).Decode(&job); err != nil {
		return Job{}, err
	}
	return job, nil
}
func tryReload(task *Task, job *Job) error {

	if err := pullReop(*task); err != nil {

		fmt.Println("Error in pulling repo:", task.RepoURL, err) //task.RepoURL
		return err
	}

	newjob, err := getJob(*task)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		log.Println(err)
		return err
	}
	if newjob.Interval != *&job.Interval || newjob.Script != *&job.Script {
		log.Println("Job updated")
		*job = newjob
		crond.Stop()
		crond.Remove(cron.EntryID(0))
		cron.DelayIfStillRunning(nil) //fixme
		crond.AddFunc(job.Interval, jobFunc)
		crond.Start()
	}
	// verbose:
	// else {
	// 	fmt.Println("Job not updated")
	// }
	return nil
}
func createCronFunc(task *Task, job *Job, oneshot bool) func() {

	return func() {
		if HACKLOCK_is_running {
			//verbose:
			fmt.Println("Job already running")
			return
		}
		HACKLOCK_is_running = true
		defer func() {
			HACKLOCK_is_running = false
			//verbose:
			fmt.Println("hacklock released")
		}()
		runid := uuid.New().String()
		fmt.Println("Running job", task.Name, "with id", runid)
		defer fmt.Println("Finish job", task.Name, "with id", runid)

		rev := getRevision(*task)
		if triggeronchange && rev != "" {
			if rev == getSavedRevision(*task) {
				//verbose:
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
