package model

import (
	"encoding/json"
	"fmt"
	"utils"
)

type Language int8

const (
	LanguageC    Language = 0
	LanguageCpp           = 1
	LanguageJava          = 2
	LanguageGo            = 3
)

type JudgeStatus int8

const (
	JudgeStatusSystemError                  JudgeStatus = -1
	JudgeStatusWaiting                                  = 0
	JudgeStatusCompiling                                = 1
	JudgeStatusCompilationError                         = 2
	JudgeStatusCompilationTimeLimitExceeded             = 3
	JudgeStatusRunning                                  = 4
	JudgeStatusTimeLimitExceeded                        = 5
	JudgeStatusMemoryLimitExceeded                      = 6
	JudgeStatusOutputLimitExceeded                      = 7
	JudgeStatusRuntimeError                             = 8
	JudgeStatusPresentationError                        = 9
	JudgeStatusWrongAnswer                              = 10
	JudgeStatusAccept                                   = 11
	JudgeStatusWaitingRunning                           = 12
)

type MissionModel struct {
	Rid         int64    `json:"rid"`
	Pid         int64    `json:"pid"`
	Code        string   `json:"code"`
	Language    Language `json:"language"`
	TimeLimit   int      `json:"time_limit"`
	MemoryLimit int      `json:"memory_limit"`
	//currentCase int64
	//caseCount   int64
}

func (m *MissionModel) LogFetchSuccess() {

}

func (m *MissionModel) LogDispatchSuccess() {

}

type StatusModel struct {
	Rid                int64       `json:"rid"`
	Pid                int64       `json:"pid"`
	Status             JudgeStatus `json:"status"`
	TimeCost           int64       `json:"time_cost,omitempty"`
	MemoryCost         int64       `json:"memory_cost,omitempty"`
	CompilationMessage string      `json:"compilation_message,omitempty"`
	//Percent float32     `json:"percent"`
}

func (s StatusModel) MarshalJSON() ([]byte, error) {
	type Alias StatusModel
	var timeCost, memoryCost *int64
	if s.TimeCost >= 0 {
		timeCost = &s.TimeCost
	}
	if s.MemoryCost >= 0 {
		memoryCost = &s.MemoryCost
	}

	return json.Marshal(&struct {
		Alias
		TimeCost   *int64 `json:"time_cost,omitempty"`
		MemoryCost *int64 `json:"memory_cost,omitempty"`
	}{
		Alias:      Alias(s),
		TimeCost:   timeCost,
		MemoryCost: memoryCost,
	})
}

func (s *StatusModel) LogTrySend() {
	if s.Status == JudgeStatusAccept {
		utils.Log(utils.LogTypeNormal, fmt.Sprintf("[Rid:%d] [Status:%d] [TimeCost:%d] [MemoryCost:%d] try send", s.Rid, s.Status, s.TimeCost, s.MemoryCost))
	} else {
		utils.Log(utils.LogTypeNormal, fmt.Sprintf("[Rid:%d] [Status:%d] try send", s.Rid, s.Status))
	}
}

func (s *StatusModel) LogSendSuccess() {
	if s.Status == JudgeStatusAccept {
		utils.Log(utils.LogTypeNormal, fmt.Sprintf("[Rid:%d] [Status:%d] [TimeCost:%d] [MemoryCost:%d] send success", s.Rid, s.Status, s.TimeCost, s.MemoryCost))
	} else {
		utils.Log(utils.LogTypeNormal, fmt.Sprintf("[Rid:%d] [Status:%d] send success", s.Rid, s.Status))
	}
}
