package utils

import (
	"config"
	"fmt"
	"os"
	"time"
)

type LogType string

const (
	LogTypeNormal  LogType = "Normal"
	LogTypeWarning LogType = "Warning"
	LogTypeError   LogType = "Error"
)

func Log(logType LogType, log string) {
	content := fmt.Sprintf("[%s] [%s]: %s\n", time.Now().Format("2006-01-02 15:04:05"), logType, log)
	logFile, err := os.OpenFile(config.GetLogFilePath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer func(logFile *os.File) {
		_ = logFile.Close()
	}(logFile)
	_, _ = logFile.WriteString(content)
	fmt.Println(content)
}
