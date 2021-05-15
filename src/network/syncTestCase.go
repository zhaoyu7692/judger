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

func init() {
	serverConfig := config.GlobalConfig.Server
	downloadURL = fmt.Sprintf("http://%s:%s/downlaodTestCase?filename=", serverConfig.Host, serverConfig.Port)
	checkURL = fmt.Sprintf("http://%s:%s/checkTestCase", serverConfig.Host, serverConfig.Port)
}

type syncTestCaseRequestModel struct {
	Filenames []string `json:"filenames"`
	Pid       int64    `json:"pid"`
}

type syncTestCaseResponseModel struct {
	Filenames []string `json:"filenames"`
}

func SyncTestCaseWithPid(pid int64) bool {
	client := http.Client{Timeout: 30 * time.Second}
	testCases, err := ioutil.ReadDir(config.GlobalConfig.Path.Data + strconv.FormatInt(pid, 10))
	if err != nil {
		//m.LogError("open testcase directory fail")
		return false
	}
	//m.LogNormal("open testcase directory success")
	requestModel := syncTestCaseRequestModel{
		Pid: pid,
	}
	for _, file := range testCases {
		inputFileRegex := regexp.MustCompile("^.+\\.in$")
		outputFileRegex := regexp.MustCompile("^.+\\.out$")
		if inputFileRegex.MatchString(file.Name()) || outputFileRegex.MatchString(file.Name()) {
			requestModel.Filenames = append(requestModel.Filenames, file.Name())
		}
	}
	data, err := json.Marshal(requestModel)
	if err != nil {
		return false
	}
	if response, err := client.Post(checkURL, "application/json", bytes.NewBuffer(data)); err == nil {
		if body, err := ioutil.ReadAll(response.Body); err == nil {
			responseModel := syncTestCaseResponseModel{}
			if err := json.Unmarshal(body, &responseModel); err == nil {
				if len(responseModel.Filenames) == 0 {
					return true
				}
				return downloadFiles(pid, &responseModel)
			}
		}
	}
	return false
}

func downloadFiles(pid int64, model *syncTestCaseResponseModel) bool {
	for _, filename := range model.Filenames {
		response, err := http.Get(downloadURL + filename)
		if err != nil {
			return false
		}
		data, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return false
		}
		savePath := fmt.Sprintf("%s%d/%s", config.GlobalConfig.Path.Data, pid, filename)
		err = ioutil.WriteFile(savePath, data, os.ModePerm)
		if err != nil {
			return false
		}
	}
	return true
}
