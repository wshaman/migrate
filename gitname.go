package migrate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
)

func WhoAmI() (name, email string) {
	var ok bool
	s, err := readCli("git", "--version")
	if err != nil || s == "" {
		return
	}
	if name, email, ok = askName("local"); ok {
		return
	}
	name, email, _ = askName("global")
	return strings.Trim(name, "\n "), strings.Trim(email, "\n ")
}

func askName(target string) (name, email string, ok bool) {
	var err error
	if name, err = readCli("git", "config", "--"+target, "--get", "user.name"); err != nil {
		return
	}
	if email, err = readCli("git", "config", "--"+target, "--get", "user.email"); err != nil {
		return
	}
	return
}

func readCli(name string, arg ...string) (string, error) {
	b, err := cliExec(name, arg...)
	if err != nil {
		return "", err
	}
	g, err := ioutil.ReadAll(b)
	if err != nil {
		return "", err
	}
	return string(g), nil
}

func cliExec(name string, arg ...string) (*bytes.Buffer, error) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, name, arg...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := err.Error()
		if stderr.Len() > 0 {
			errMsg = stderr.String()
		}
		return nil, errors.New(fmt.Sprintf("%s %s error: %s", name, strings.Join(arg, " "), errMsg))
	}
	return &stdout, nil
}
