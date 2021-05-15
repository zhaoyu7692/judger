package machine

import "os/exec"

type CMachine struct {
	BaseMachine
}

func (c *CMachine) compileCommand() *exec.Cmd {
	return exec.Command("gcc", "-g", "-Wall", "-o", "main", "main.c")
}

func (c *CMachine) judgeCommand() *exec.Cmd {
	return exec.Command("./main")
}

func (c *CMachine) sourceCodeFileName() string {
	return "main.c"
}
