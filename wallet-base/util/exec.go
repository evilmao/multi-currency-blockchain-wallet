package util

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// ExeCmd is a simple wrapper of exec.Cmd.
func ExeCmd(name string, args []string, setter func(*exec.Cmd)) ([]byte, error) {
	c := exec.Command(name, args...)
	if setter != nil {
		setter(c)
	}
	return c.CombinedOutput()
}

func NoHup(cmd string, args []string, setter func(*exec.Cmd)) ([]byte, error) {
	cmd = fmt.Sprintf("nohup %s %s > /dev/null 2> /dev/null < /dev/null &", cmd, strings.Join(args, " "))
	return ExeCmd("bash", []string{"-c", cmd}, setter)
}

func SetSID(cmd string, args []string, setter func(*exec.Cmd)) ([]byte, error) {
	cmd = fmt.Sprintf("setsid %s %s > /dev/null 2> /dev/null < /dev/null &", cmd, strings.Join(args, " "))
	return ExeCmd("bash", []string{"-c", cmd}, setter)
}

func KillName(name string) error {
	pid, err := GetPID(name)
	if err != nil {
		return fmt.Errorf("get %s pid failed, %v", name, err)
	}

	err = KillPID(pid)
	if err != nil {
		return fmt.Errorf("kill pid %d failed, %v", pid, err)
	}

	const maxTry = 10
	for i := 0; i < maxTry; i++ {
		time.Sleep(time.Second)

		pid, _ := GetPID(name)
		if pid <= 0 {
			return nil
		}
	}

	return fmt.Errorf("kill %s(%d) timeout", name, pid)
}

func GetPID(name string) (int, error) {
	out, _ := ExeCmd("pidof", []string{name}, nil)
	s := strings.TrimSpace(string(out))
	if len(s) == 0 {
		return 0, nil
	}

	pid, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("convert pid failed, %v", err)
	}

	return pid, nil
}

func KillPID(pid int) error {
	if pid <= 0 {
		return nil
	}

	_, err := ExeCmd("kill", []string{strconv.Itoa(pid)}, nil)
	return err
}
