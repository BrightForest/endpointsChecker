package main

import (
	"fmt"
	"log"
	"gopkg.in/yaml.v2"
	"os"
	"io/ioutil"
	"path/filepath"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var Ss = map[string]int{}

type CheckInstance struct {
	CheckName string `yaml:"checkName"`
	CheckEndpoint string `yaml:"checkEndpoint"`
	CheckRateSeconds int `yaml:"checkRateSeconds"`
	CheckSuccessString string `yaml:"checkSuccessString"`
	CheckType string `yaml:"checkType"`
}

type Configs struct {
	Cfgs []CheckInstance `tasks`
}

func main() {
	//Get config
	confFilePath, _ := filepath.Abs(os.Getenv("CHECKER_CONF_FILE"))
	yamlFile, err := ioutil.ReadFile(confFilePath)

	if err != nil {
		fmt.Println("Config file not found " + confFilePath)
		panic(err)
	}
	//Parse yaml
	var config Configs

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	checkSheduler(config)
}

func checkSheduler(config Configs) {
	for i := 0; i < len(config.Cfgs); i++ {
		if (strings.Contains("http", config.Cfgs[i].CheckType)) {
			go checkHttpService(config.Cfgs[i].CheckEndpoint, config.Cfgs[i].CheckSuccessString, config.Cfgs[i].CheckRateSeconds,config.Cfgs[i].CheckName)
		}
	}
	go webMetricsService()
	for {
		time.Sleep(1000 * time.Millisecond * time.Duration(10))
	}
}

func checkHttpService(checkEndpoint string, checkSuccessString string, checkRateSeconds int, checkName string) {
	var returnState int
	checkSuccessInt, err := strconv.Atoi(checkSuccessString)
	if (err != nil) {
		fmt.Println("Unable to convert string to int success state for endpoint: " + checkEndpoint)
	}
	for {
		time.Sleep(1000 * time.Millisecond * time.Duration(checkRateSeconds))
		resp, err := http.Get(checkEndpoint)
		if (err == nil) && (resp.StatusCode == checkSuccessInt) {
			returnState = 1
			Ss[checkName + "_http_state_up"] = returnState
		} else {
			returnState = 0
			Ss[checkName + "_http_state_up"] = returnState
		}
	}
}

func webMetricsService() {
	http.HandleFunc("/metrics", handler)
	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	for key, value := range Ss {
		fmt.Fprintln(w, key, value)
	}
}