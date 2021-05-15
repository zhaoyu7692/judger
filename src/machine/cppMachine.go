package machine

import "os/exec"

type CppMachine struct {
	BaseMachine
}

func (c *CppMachine) compileCommand() *exec.Cmd {
	return exec.Command("g++", "-g", "-Wall", "-o", "main", "main.cpp")
}

func (c *CppMachine) judgeCommand() *exec.Cmd {
	return exec.Command("./main")
}

func (c *CppMachine) sourceCodeFileName() string {
	return "main.cpp"
}
