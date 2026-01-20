package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func installCurr() {
	buildTime := time.Now().Unix()
	commitHash := strings.TrimSpace(runCmd("git", "rev-parse", "--short", "HEAD"))
	commitMsg := strings.TrimSpace(runCmd("git", "log", "-1", "--pretty=%B"))
	commitMsg = strings.TrimSpace(strings.ReplaceAll(commitMsg, "\n", " "))
	commitMsg = base64.StdEncoding.EncodeToString([]byte(commitMsg))

	ldflags := fmt.Sprintf(
		"-X 'main.buildTime=%d' -X 'main.gitCommit=%s' -X 'main.gitCommitMsg=%s' -s -w",
		buildTime, commitHash, commitMsg)

	buildDir := "build"
	progName := "mujamalat"
	progDest := filepath.Join(buildDir, progName)
	runCmd("go", "build", "-ldflags", ldflags, "-tags", "static netgo", "-o", progDest)

	installDir := os.Getenv("BIN_DIR")
	if installDir == "" {
		if termuxPre := os.Getenv("TERMUX__PREFIX"); termuxPre != "" {
			installDir = filepath.Join(termuxPre, "bin")
		} else {
			home := must(os.UserHomeDir())
			installDir = filepath.Join(home, ".local/bin")
		}
	}

	installDest := filepath.Join(installDir, progName)
	os.Remove(installDest)
	must(0, os.Rename(progDest, installDest))
	fmt.Printf("Installed: %s\n", installDest)
}

func runCmd(c string, a ...string) string {
	buf := bytes.Buffer{}

	cmd := exec.Command(c, a...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		fmt.Println(buf.String())
		panic(err)
	}

	return buf.String()
}

func runCmdEcho(c string, a ...string) {
	cmd := exec.Command(c, a...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()

	if err != nil {
		panic(err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
