package network

import (
	"bytes"
	"config"
	"encoding/json"
	"io/ioutil"
	"model"
	"net/http"
	"sync"
	"time"
	"utils"
)

var url string
var judgingCount int
var statusList []model.StatusModel
var missionList []model.MissionModel
var countLock sync.Mutex
var statusLock sync.Mutex
var missionLock sync.Mutex

type MissionRequestModel struct {
	Status       []model.StatusModel `json:"status"`
	JudgingCount int                 `json:"judging_count"`
}

type MissionResponseModel struct {
	Problems []model.MissionModel `json:"problems"`
}

func init() {
	judgingCount = 0
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
		countLock.Unlock()
	}
	missionLock.Unlock()
	return mission
}

func SendStatus(statusModel model.StatusModel) {
	statusLock.Lock()
	statusList = append(statusList, statusModel)
	switch statusModel.Status {
	case model.JudgeStatusSystemError:
	case model.JudgeStatusCompilationError:
	case model.JudgeStatusCompilationTimeLimitExceeded:
	case model.JudgeStatusTimeLimitExceeded:
	case model.JudgeStatusMemoryLimitExceeded:
	case model.JudgeStatusOutputLimitExceeded:
	case model.JudgeStatusRuntimeError:
	case model.JudgeStatusPresentationError:
	case model.JudgeStatusWrongAnswer:
	case model.JudgeStatusAccept:
		countLock.Lock()
		judgingCount--
		countLock.Unlock()
	}
	statusLock.Unlock()
}

func LogNormal(content string) {
	utils.Log(utils.LogTypeNormal, content)
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
				missionLock.Lock()
				for _, mission := range responseModel.Problems {
					missionList = append(missionList, mission)
				}
				missionLock.Unlock()
			}
		}
		_ = response.Body.Close()
	}
}

func StartNetworkModule() {
	for i := 0; ; i++ {
		fetchMissionAndSendStatus()
		time.Sleep(time.Second)
	}
}
