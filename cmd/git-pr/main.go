package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// exit handles the common error exit case
func exit(err error) {
	if exitErr, ok := err.(*exec.ExitError); ok {
		fmt.Println(exitErr.Stderr)
	}

	os.Exit(1)
}

// run runs the given command with stdout and stderr mapped to this process
func run(program string, args ...string) error {
	cmd := exec.Command(program, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// head returns the current head for the repo
func head() string {
	headRef, headErr := exec.Command(
		"git",
		"rev-parse",
		"--abbrev-ref",
		"HEAD",
	).Output()

	if headErr != nil {
		exit(headErr)
	}

	return strings.TrimSpace(string(headRef))
}

// status returns the current stats of the repo
func status() string {
	statusOut, statusErr := exec.Command(
		"git",
		"status",
		"-s",
	).Output()

	if statusErr != nil {
		exit(statusErr)
	}

	return string(statusOut)
}

// changes returns the number of lines in the status output
func changes() int {
	statusOut := status()

	changes := strings.Split(statusOut, "\n")
	count := 0

	for _, line := range changes {
		line = strings.TrimSpace(line)
		if line != "" {
			count++
		}
	}

	return count
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: git share [name] [rev]")
		os.Exit(1)
	}

	name := os.Args[1]
	rev := os.Args[2]

	headRev := head()
	changeCount := changes()

	if changeCount > 0 {
		fmt.Println("Cannot continue: pending changes")
		os.Exit(1)
	}

	// create the new branch
	branchErr := run(
		"git",
		"branch",
		name,
		"origin/master",
		"--no-track",
	)
	if branchErr != nil {
		os.Exit(1)
	}

	// check out the new branch
	checkoutErr := run(
		"git",
		"checkout",
		"-q",
		name,
	)
	if checkoutErr != nil {
		os.Exit(1)
	}

	// cherry pick
	cherrypickErr := run(
		"git",
		"cherry-pick",
		rev,
	)
	if cherrypickErr != nil {
		os.Exit(1)
	}

	// push
	pushErr := run(
		"git",
		"push",
		"origin",
		name,
	)
	if pushErr != nil {
		os.Exit(1)
	}

	// check out original head
	checkoutOrigErr := run(
		"git",
		"checkout",
		"-q",
		headRev,
	)
	if checkoutOrigErr != nil {
		os.Exit(1)
	}

	// delete the branch that was created
	deleteBranchErr := run(
		"git",
		"branch",
		"-D",
		name,
	)
	if deleteBranchErr != nil {
		os.Exit(1)
	}
}