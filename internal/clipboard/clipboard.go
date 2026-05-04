package clipboard

import (
	"errors"
	"os/exec"
)

var commands = [][]string{
	{"wl-copy"},
	{"xclip", "-selection", "clipboard"},
	{"xsel", "--clipboard", "--input"},
}

func Available() bool {
	_, err := command()
	return err == nil
}

func Copy(text string) error {
	cmdArgs, err := command()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	_, _ = stdin.Write([]byte(text))
	_ = stdin.Close()
	return cmd.Wait()
}

func command() ([]string, error) {
	for _, candidate := range commands {
		if _, err := exec.LookPath(candidate[0]); err == nil {
			return candidate, nil
		}
	}
	return nil, errors.New("no clipboard command found: install wl-clipboard, xclip, or xsel")
}
