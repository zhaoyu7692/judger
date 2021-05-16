package network

import (
	"bytes"
	"config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"model"
	"net/http"
	"sync"
	"time"
	"utils"
)

var url string

var judgingCount int
var judgingStatus map[int64]int64
var countLock sync.Mutex

var statusList []model.StatusModel
var statusLock sync.Mutex

var missionList []model.MissionModel
var missionLock sync.Mutex

var lockedMission map[int64][]model.MissionModel
var lockedMissionLock sync.Mutex

var syncingPid map[int64]bool

type MissionRequestModel struct {
	Status       []model.StatusModel `json:"status"`
	JudgingCount int                 `json:"judging_count"`
}

type MissionResponseModel struct {
	Problems []model.MissionModel `json:"problems"`
}

func init() {
	judgingCount = 0
	syncingPid = map[int64]bool{}
	judgingStatus = map[int64]int64{}
	lockedMission = map[int64][]model.MissionModel{}
	url = "http://" + config.GlobalConfig.Server.Host + ":" + config.GlobalConfig.Server.Port + "/api/core/j2s/"
}

func FetchMission() *model.MissionModel {
	var mission *model.MissionModel
	missionLock.Lock()
	if len(missionList) > 0 {
		mission = &missionList[0]
		missionList = missionList[1:]
		countLock.Lock()
		judgingCount++
		count := judgingStatus[mission.Pid]
		judgingStatus[mission.Pid] = count + 1
		countLock.Unlock()
	}
	missionLock.Unlock()
	return mission
}

func SendStatus(statusModel model.StatusModel) {
	statusLock.Lock()
	statusList = append(statusList, statusModel)
	switch statusModel.Status {
	case model.JudgeStatusSystemError,
		model.JudgeStatusCompilationError,
		model.JudgeStatusCompilationTimeLimitExceeded,
		model.JudgeStatusTimeLimitExceeded,
		model.JudgeStatusMemoryLimitExceeded,
		model.JudgeStatusOutputLimitExceeded,
		model.JudgeStatusRuntimeError,
		model.JudgeStatusPresentationError,
		model.JudgeStatusWrongAnswer,
		model.JudgeStatusAccept:
		countLock.Lock()
		judgingCount--
		count := judgingStatus[statusModel.Pid]
		judgingStatus[statusModel.Pid] = count - 1
		if judgingStatus[statusModel.Pid] == 0 {
			SyncTestCase(statusModel.Pid)
		}
		countLock.Unlock()
	}
	statusLock.Unlock()
}

func LogNormal(pid int64, content string) {
	utils.Log(utils.LogTypeNormal, fmt.Sprintf("[Pid:%d] %s", pid, content))
}
func LogWarning(pid int64, content string) {
	utils.Log(utils.LogTypeWarning, fmt.Sprintf("[Pid:%d] %s", pid, content))
}
func LogError(pid int64, content string) {
	utils.Log(utils.LogTypeError, fmt.Sprintf("[Pid:%d] %s", pid, content))
}

func fetchMissionAndSendStatus() {
	statusLock.Lock()
	sendCount := utils.Min(len(statusList), 20)
	for i := 0; i < sendCount; i++ {
		statusList[i].LogTrySend()
	}
	requestModel := MissionRequestModel{
		Status:       statusList[:sendCount],
		JudgingCount: judgingCount,
	}
	statusLock.Unlock()
	client := http.Client{Timeout: 30 * time.Second}
	data, _ := json.Marshal(requestModel)
	response, err := client.Post(url, "application/json", bytes.NewBuffer(data))
	if err == nil {
		body, err := ioutil.ReadAll(response.Body)
		if err == nil {
			var responseModel MissionResponseModel
			err := json.Unmarshal(body, &responseModel)
			if err == nil {
				statusLock.Lock()
				for i := 0; i < sendCount; i++ {
					statusList[i].LogSendSuccess()
				}
				statusList = statusList[sendCount:]
				statusLock.Unlock()
				for _, mission := range responseModel.Problems {
					if needSync, _ := CheckTestCaseWithPid(mission.Pid); needSync == true {
						lockedMissionLock.Lock()
						lockedMission[mission.Pid] = append(lockedMission[mission.Pid], mission)
						lockedMissionLock.Unlock()
						if judgingStatus[mission.Pid] == 0 {
							SyncTestCase(mission.Pid)
						}
					} else {
						missionLock.Lock()
						missionList = append(missionList, mission)
						missionLock.Unlock()
					}
				}
			}
		}
		_ = response.Body.Close()
	}
}

var syncLock sync.Mutex

func SyncTestCase(pid int64) {
	syncLock.Lock()
	defer syncLock.Unlock()
	if syncingPid[pid] {
		return
	}
	syncingPid[pid] = true

	go SyncTestCaseWithPid(pid, func() {
		lockedMissionLock.Lock()
		missionLock.Lock()
		for _, mission := range lockedMission[pid] {
			missionList = append(missionList, mission)
		}
		lockedMission[pid] = lockedMission[pid][0:0]
		missionLock.Unlock()
		lockedMissionLock.Unlock()
		syncLock.Lock()
		syncingPid[pid] = true
		syncLock.Unlock()
	})
}

func StartNetworkModule() {
	for i := 0; ; i++ {
		fetchMissionAndSendStatus()
		time.Sleep(time.Second)
	}
}
