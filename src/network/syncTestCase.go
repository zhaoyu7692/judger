package network

import (
	"bytes"
	"config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var downloadURL, checkURL string
var missions map[int64]bool

func init() {
	missions = map[int64]bool{}
	serverConfig := config.GlobalConfig.Server
	downloadURL = fmt.Sprintf("http://%s:%s/downloadTestCase/", serverConfig.Host, serverConfig.Port)
	checkURL = fmt.Sprintf("http://%s:%s/checkTestCase/", serverConfig.Host, serverConfig.Port)
}

type syncTestCaseRequestModel struct {
	Filenames []string `json:"filenames"`
	Pid       int64    `json:"pid"`
}

type syncTestCaseResponseModel struct {
	Filenames       []string `json:"filenames"`
	RemoveFilenames []string `json:"remove_filenames"`
}

func CheckTestCaseWithPid(pid int64) (bool, *syncTestCaseResponseModel) {
	LogNormal(pid, "start check test case")
	success := false
	defer func() {
		LogNormal(pid, fmt.Sprintf("[NeedSync:%d] check test case complete", success))
	}()
	client := http.Client{Timeout: 30 * time.Second}
	testCases, err := ioutil.ReadDir(config.GlobalConfig.Path.Data + strconv.FormatInt(pid, 10))
	if err != nil {
		LogError(pid, "open testcase directory fail")
		return false, nil
	}
	requestModel := syncTestCaseRequestModel{
		Pid: pid,
	}
	inputFileRegex := regexp.MustCompile("^.+\\.in$")
	outputFileRegex := regexp.MustCompile("^.+\\.out$")
	for _, file := range testCases {
		if inputFileRegex.MatchString(file.Name()) || outputFileRegex.MatchString(file.Name()) {
			requestModel.Filenames = append(requestModel.Filenames, file.Name())
		}
	}
	data, err := json.Marshal(requestModel)
	if err != nil {
		return false, nil
	}
	if response, err := client.Post(checkURL, "application/json", bytes.NewBuffer(data)); err == nil {
		if body, err := ioutil.ReadAll(response.Body); err == nil {
			responseModel := syncTestCaseResponseModel{}
			if err := json.Unmarshal(body, &responseModel); err == nil {
				if len(responseModel.Filenames) > 0 {
					missions[pid] = true
				}
				success = true
				return true, &responseModel
			}
		}
	}
	return false, nil
}

func SyncTestCaseWithPid(pid int64, callback func()) {
	if missions[pid] == false {
		callback()
	}
	LogNormal(pid, "start sync test case")
	warredCallback := func() {
		LogNormal(pid, "sync test case complete")
		callback()
	}
	if needSync, model := CheckTestCaseWithPid(pid); needSync == true {
		for _, filename := range model.Filenames {
			LogNormal(pid, "download "+filename)
			response, err := http.Get(fmt.Sprintf("%s?pid=%d&filename=%s", downloadURL, pid, filename))
			if err != nil || response.StatusCode == 404 {
				time.Sleep(30 * time.Second)
				LogWarning(pid, fmt.Sprintf("[filename:%s] download fail, retry after 30 senonds", filename))
				SyncTestCaseWithPid(pid, callback)
				return
			}
			data, err := ioutil.ReadAll(response.Body)
			if err != nil {
				time.Sleep(30 * time.Second)
				LogWarning(pid, fmt.Sprintf("[filename:%s] download fail, retry after 30 senonds", filename))
				SyncTestCaseWithPid(pid, callback)
				return
			}
			savePath := fmt.Sprintf("%s%d/%s", config.GlobalConfig.Path.Data, pid, filename)
			err = ioutil.WriteFile(savePath, data, os.ModePerm)
			if err != nil {
				LogError(pid, fmt.Sprintf("[filename:%s] save fail, retry after 30 senonds", filename))
				time.Sleep(30 * time.Second)
				SyncTestCaseWithPid(pid, callback)
				return
			}
			LogNormal(pid, "download "+filename+"complete")
		}
		for _, filename := range model.RemoveFilenames {
			removePath := fmt.Sprintf("%s%d/%s", config.GlobalConfig.Path.Data, pid, filename)
			if err := os.Remove(removePath); err != nil {
				time.Sleep(30 * time.Second)
				LogWarning(pid, fmt.Sprintf("[filename:%s] remove fail, retry after 30 senonds", filename))
				SyncTestCaseWithPid(pid, callback)
				return
			}
		}
		LogNormal(pid, "sync test case complete")
		warredCallback()
	}
	return
}
