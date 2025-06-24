package gitaskop

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
)

const (
	tmpFolder = "gitaskop"
)

func cloneRepo(task Task, force bool) error {
	folder := getRepoFolder(task)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		fmt.Println("cloning repo (" + task.Branch + ")")
		cmd := exec.Command("git", "clone", "--branch", task.Branch, task.RepoURL, folder)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to clone repo:", err)
			return err
		}

		fmt.Println("Repo cloned successfully")
		return nil
	} else if force {
		err := cleanRepo(task)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to clean repo:", err)
			return err
		}
		return cloneRepo(task, false)
	}

	fmt.Println("Repo already cloned, pulling latest changes")
	return pullRepo(task)
}

func cleanRepo(task Task) error {
	folder := getRepoFolder(task)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(folder); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to clean repo:", err)
		return err
	}
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil
	}
	fmt.Println("Repo cleaned successfully")
	return nil
}

func resetRepo(task Task) error {
	folder := getRepoFolder(task)
	fmt.Println("resetting repo")

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return nil
	}
	cmd := exec.Command("git", "reset", "--hard", "HEAD")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = folder

	err := cmd.Run()
	cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to reset repo:", err)
		return err
	}
	return nil
}

func pullRepo(task Task) error {
	folder := getRepoFolder(task)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		return err
	}

	cmd := exec.Command("git", "pull", "--quiet")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = folder

	err := cmd.Run()
	cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}

func getRepoFolder(task Task) string {
	ps := string(os.PathSeparator)
	hash := md5.Sum([]byte(task.RepoURL))
	return os.TempDir() + ps + tmpFolder + ps + hex.EncodeToString(hash[:])
}

func getRevision(task Task) string {
	folder := getRepoFolder(task)
	cmd := exec.Command("git", "-C", folder, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get revision:", err)
		return ""
	}
	return string(out)
}

func saveRevision(task Task) {
	folder := getRepoFolder(task)
	revision := getRevision(task)
	if revision == "" {
		return
	}
	file, err := os.Create(folder + "_revision")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save revision:", err)
		return
	}
	defer file.Close()
	file.WriteString(revision)
}

func getSavedRevision(task Task) string {
	folder := getRepoFolder(task)
	file, err := os.Open(folder + "_revision")
	if err != nil {
		return ""
	}
	defer file.Close()
	buf := make([]byte, 64)
	n, err := file.Read(buf)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read revision:", err)
		return ""
	}
	return string(buf[:n])
}

func hasRepoChanged(task Task) bool {
	return getRevision(task) != getSavedRevision(task)
}