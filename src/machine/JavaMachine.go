package machine

import "os/exec"

type JavaMachine struct {
	BaseMachine
}

func (c *JavaMachine) compileCommand() *exec.Cmd {
	return exec.Command("javac", "Main.java")
}

func (c *JavaMachine) judgeCommand() *exec.Cmd {
	return exec.Command("java", "Main")
}

func (c *JavaMachine) sourceCodeFileName() string {
	return "Main.java"
}
