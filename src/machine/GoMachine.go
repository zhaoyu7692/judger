package machine

import "os/exec"

type GoMachine struct {
	BaseMachine
}

func (c *GoMachine) compileCommand() *exec.Cmd {
	return exec.Command("go", "build", "-o", "main", "main.go")
}

func (c *GoMachine) judgeCommand() *exec.Cmd {
	return exec.Command("./main")
}

func (c *GoMachine) sourceCodeFileName() string {
	return "main.go"
}
