package machine

import (
	"config"
	"fmt"
	"io/ioutil"
	"model"
	"network"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"utils"
)

type Machine interface {
	compileCommand() *exec.Cmd
	judgeCommand() *exec.Cmd
	sourceCodeFileName() string
	Run(machine Machine)
}

type BaseMachine struct {
	Rid         int64             `json:"rid"`
	Pid         int64             `json:"pid"`
	Code        string            `json:"code"`
	Language    model.Language    `json:"language"`
	Status      model.JudgeStatus `json:"status"`
	TimeLimit   int               `json:"time_limit"`
	MemoryLimit int               `json:"memory_limit"`

	timeCost           int
	memoryCost         int
	compilationMessage string
	inputFiles         []string
	//currentCase int64
	//caseCount   int64
}

func (m BaseMachine) LogNormal(content string) {
	utils.Log(utils.LogTypeNormal, fmt.Sprintf("[Rid:%d] [Pid:%d]: %s", m.Rid, m.Pid, content))
}

func (m BaseMachine) LogWarning(content string) {
	utils.Log(utils.LogTypeWarning, fmt.Sprintf("[Rid:%d] [Pid:%d]: %s", m.Rid, m.Pid, content))
}

func (m BaseMachine) LogError(content string) {
	utils.Log(utils.LogTypeError, fmt.Sprintf("[Rid:%d] [Pid:%d]: %s", m.Rid, m.Pid, content))
}

func (m *BaseMachine) workPath() string {
	return config.GlobalConfig.Path.Work + strconv.FormatInt(m.Rid, 10)
}

func (m *BaseMachine) dataPath() string {
	return config.GlobalConfig.Path.Data + strconv.FormatInt(m.Pid, 10)
}

func (m *BaseMachine) initWorkSpace(machine Machine) {
	m.LogNormal("start initializing workspace")
	// calculate test case count
	m.LogNormal(fmt.Sprintf("open testcase directory %s", m.dataPath()))
	testCases, err := ioutil.ReadDir(m.dataPath())
	if err != nil {
		m.Status = model.JudgeStatusSystemError
		m.LogError("open testcase directory fail")
		return
	}
	m.LogNormal("open testcase directory success")
	for _, file := range testCases {
		inputFileRegex := regexp.MustCompile("^.+\\.in$")
		if result := inputFileRegex.MatchString(file.Name()); result == true {
			_, err := os.Open(fmt.Sprintf("%s/%s.out", m.dataPath(), file.Name()[:strings.LastIndex(file.Name(), ".in")]))
			if err == nil {
				m.inputFiles = append(m.inputFiles, file.Name())
			}
		}
	}
	m.LogNormal(fmt.Sprintf("find %d testcase(s)", len(m.inputFiles)))

	// create work directory
	m.LogNormal(fmt.Sprintf("detect work directory %s existance", m.workPath()))
	err = syscall.Access(m.workPath(), 0)
	if err != nil {
		m.LogNormal("work directory not exist")
		m.LogNormal("create work directory")
		err := os.MkdirAll(m.workPath(), 0777)
		if err != nil {
			m.Status = model.JudgeStatusSystemError
			m.LogError("create work directory fail")
			return
		}
		m.LogNormal("create work directory success")
	}
	m.LogNormal("work directory exist")

	// save source code
	sourceCodePath := fmt.Sprintf("%s%d/%s", config.GlobalConfig.Path.Work, m.Rid, machine.sourceCodeFileName())
	m.LogNormal("save source code to " + sourceCodePath)
	err = ioutil.WriteFile(sourceCodePath, []byte(m.Code), os.ModePerm)
	if err != nil {
		m.Status = model.JudgeStatusSystemError
		m.LogError("save source code fail")
		return
	}
	m.LogNormal("save source code success")

	m.LogNormal("workspace initialization complete")
}

func (m *BaseMachine) compile(machine Machine) {
	m.LogNormal("start compile source code")
	cmd := machine.compileCommand()
	cmd.Stdin = os.Stdin
	// redirect compile message output stream
	m.LogNormal("create compile message file")
	compileMessageFile, err := os.Create(m.workPath() + "/compile.log")
	if err != nil {
		m.LogNormal("create compile message file fail")
		compileMessageFile = os.Stdout
	} // TODO: log position
	m.LogNormal("create compile message file success")

	defer func(compileMessageFile *os.File) {
		_ = compileMessageFile.Close()
		if file, err := ioutil.ReadFile(m.workPath() + "/compile.log"); err == nil {
			m.compilationMessage = string(file)
		}
		m.LogNormal("source code compiling complete")
	}(compileMessageFile)
	cmd.Stdout = compileMessageFile
	cmd.Stderr = compileMessageFile
	cmd.Dir = m.workPath()
	go func() {
		_ = cmd.Run()
		if cmd.Process == nil {
			m.LogError("compiling fail, process not be created")
			m.Status = model.JudgeStatusSystemError
			return
		}
	}()
	for i := 0; ; i++ {
		// compile command run error
		if m.Status == model.JudgeStatusSystemError {
			return
		}
		// judge process after compile exited
		if cmd.ProcessState != nil {
			switch cmd.ProcessState.ExitCode() {
			case 0:
				{
					m.LogNormal("compiling success")
					m.Status = model.JudgeStatusWaitingRunning
				}
			case 1, 2:
				{
					m.LogNormal("compiling error")
					m.Status = model.JudgeStatusCompilationError
				}
			default:
				{
					m.LogNormal("compiling system error")
					m.Status = model.JudgeStatusSystemError
				}
			}
			return
		}
		// judge compile time limit exceeded
		if cmd.Process != nil {
			timeCost := utils.GetTimeUsed(cmd.Process.Pid)
			if timeCost > 20*1000 {
				_ = cmd.Process.Kill()
				m.LogNormal("compiling time limit exceeded")
				m.Status = model.JudgeStatusCompilationTimeLimitExceeded
				return
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func (m *BaseMachine) doJudge(machine Machine, inputFileName string) {
	m.LogNormal(fmt.Sprintf("start judge use %s", inputFileName))
	stdInputFile, err := os.Open(fmt.Sprintf("%s/%s", m.dataPath(), inputFileName))
	if err != nil {
		m.LogError("open standard input file fail")
		m.LogNormal(fmt.Sprintf("judge %s complete", inputFileName))
		return
	}
	defer func() {
		_ = stdInputFile.Close()
		m.LogNormal(fmt.Sprintf("judge %s complete", inputFileName))
	}()

	outputFile, err := os.Create(fmt.Sprintf("%s/%s.out", m.workPath(), inputFileName[:strings.LastIndex(inputFileName, ".in")]))
	if err != nil {
		m.LogError("create output file fail")
		return
	}
	defer func() {
		_ = outputFile.Close()
	}()

	cmd := machine.judgeCommand()
	cmd.Stdin = stdInputFile
	cmd.Stdout = outputFile
	cmd.Dir = m.workPath()

	go func() {
		_ = cmd.Run()
		if cmd.Process == nil {
			m.LogError("judge fail, process not be created")
			m.Status = model.JudgeStatusSystemError
		}
	}()
	timeCost, memoryCost := 0, 0
	defer func() {
		m.timeCost = utils.Max(m.timeCost, timeCost)
		m.memoryCost = utils.Max(m.memoryCost, memoryCost)
	}()

	for i := 0; ; i++ {
		// judge command run error
		if m.Status == model.JudgeStatusSystemError {
			return
		}
		// judge process after judge exited
		if cmd.ProcessState != nil {
			if cmd.ProcessState.Success() {
				m.LogNormal("judge success")
			} else {
				m.LogNormal("judge runtime error")
				m.Status = model.JudgeStatusRuntimeError
			}
			return
		}
		// judge compile time limit exceeded
		if cmd.Process != nil {
			timeCost = utils.Max(timeCost, utils.GetTimeUsed(cmd.Process.Pid))
			if timeCost > m.TimeLimit {
				_ = cmd.Process.Kill()
				m.LogNormal("judge time limit exceeded " + strconv.FormatInt(int64(timeCost), 10)+ "ms")
				m.Status = model.JudgeStatusTimeLimitExceeded
				return
			}
			memoryCost = utils.Max(memoryCost, utils.GetMemoryUsed(cmd.Process.Pid))
			if memoryCost > m.MemoryLimit {
				_ = cmd.Process.Kill()
				m.LogNormal("judge memory limit exceeded")
				m.Status = model.JudgeStatusMemoryLimitExceeded
				return
			}
			writeSize := utils.GetWriteBytes(cmd.Process.Pid)
			if writeSize > 256000000 {
				_ = cmd.Process.Kill()
				m.LogNormal("judge output limit exceeded")
				m.Status = model.JudgeStatusOutputLimitExceeded
				return
			}
		}
		time.Sleep(time.Millisecond)
	}
}

func (m *BaseMachine) compareOutputFile(inputFileName string) {
	outputPath := fmt.Sprintf("%s/%s.out", m.workPath(), inputFileName[:strings.LastIndex(inputFileName, ".in")])
	outputContent, err := ioutil.ReadFile(outputPath)
	if err != nil {
		m.LogError("output file not exist " + outputPath)
		return
	}
	stdOutputPath := fmt.Sprintf("%s/%s.out", m.dataPath(), inputFileName[:strings.LastIndex(inputFileName, ".in")])
	stdOutputContent, err := ioutil.ReadFile(stdOutputPath)
	if err != nil {
		m.LogError("output file not exist " + stdOutputPath)
		return
	}
	output := string(outputContent)
	stdOutput := string(stdOutputContent)
	for i := len(output) - 1; i >= 0; i-- {
		if output[i] != '\r' && output[i] != ' ' && output[i] != '\n' {
			output = output[:i+1]
			break
		}
	}
	for i := len(stdOutput) - 1; i >= 0; i-- {
		if stdOutput[i] != '\r' && stdOutput[i] != ' ' && stdOutput[i] != '\n' {
			stdOutput = stdOutput[:i+1]
			break
		}
	}
	output = strings.ReplaceAll(output, "\r", "")
	stdOutput = strings.ReplaceAll(stdOutput, "\r", "")
	if strings.Compare(output, stdOutput) != 0 {
		output = strings.ReplaceAll(output, " ", "")
		stdOutput = strings.ReplaceAll(stdOutput, " ", "")
		output = strings.ReplaceAll(output, "\n", "")
		stdOutput = strings.ReplaceAll(stdOutput, "\n", "")
		if strings.Compare(output, stdOutput) == 0 {
			m.LogNormal("judge presentation error")
			m.Status = model.JudgeStatusPresentationError
		} else {
			m.LogNormal("judge wrong answer")
			m.Status = model.JudgeStatusWrongAnswer
		}
	}
}

func (m *BaseMachine) judge(machine Machine) {
	m.LogNormal("start judge")
	for _, inputFileName := range m.inputFiles {
		if m.Status != model.JudgeStatusWaitingRunning && m.Status != model.JudgeStatusPresentationError {
			return
		}
		m.doJudge(machine, inputFileName)
		if m.Status == model.JudgeStatusWaitingRunning || m.Status == model.JudgeStatusPresentationError {
			m.compareOutputFile(inputFileName)
		}
	}
	if m.Status == model.JudgeStatusWaitingRunning {
		m.Status = model.JudgeStatusAccept
		m.LogNormal("judge accept")
	}
	m.LogNormal("judge complete")
}

func (m *BaseMachine) sendStatus() {
	timeCost, memoryCost := -1, -1
	if m.Status == model.JudgeStatusAccept {
		timeCost = m.timeCost
		memoryCost = m.memoryCost
	}
	network.SendStatus(model.StatusModel{
		Rid:                m.Rid,
		Pid:                m.Pid,
		Status:             m.Status,
		TimeCost:           int64(timeCost),
		MemoryCost:         int64(memoryCost),
		CompilationMessage: m.compilationMessage,
		//Percent:
	})
}

func (m *BaseMachine) Run(machine Machine) {
	m.LogNormal("mission start")
	m.Status = model.JudgeStatusCompiling
	m.timeCost = -1
	m.memoryCost = -1
	m.sendStatus()
	m.initWorkSpace(machine)
	m.compile(machine)
	m.sendStatus()
	if m.Status == model.JudgeStatusWaitingRunning {
		m.judge(machine)
		m.sendStatus()
	}

	m.LogNormal("mission complete")
}
