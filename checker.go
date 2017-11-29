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
	"encoding/json"
)

var Ss = map[string]int{}

type CheckInstance struct {
	CheckName string `yaml:"Name"`
	CheckEndpoint string `yaml:"Endpoint"`
	CheckRateSeconds int `yaml:"RateSeconds"`
	CheckSuccessString string `yaml:"SuccessString"`
	CheckType string `yaml:"Type"`
	CheckTimeout int `yaml:"Timeout"`
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
			go checkHttpService(
				config.Cfgs[i].CheckEndpoint,
				config.Cfgs[i].CheckSuccessString,
				config.Cfgs[i].CheckRateSeconds,
				config.Cfgs[i].CheckName,
				config.Cfgs[i].CheckTimeout)
		}
		if (strings.Contains("rest", config.Cfgs[i].CheckType)) {
			go checkJsonService(
				config.Cfgs[i].CheckEndpoint,
				config.Cfgs[i].CheckSuccessString,
				config.Cfgs[i].CheckRateSeconds,
				config.Cfgs[i].CheckName,
				config.Cfgs[i].CheckTimeout)
		}
	}
	go webMetricsService()
	for {
		time.Sleep(1000 * time.Millisecond * time.Duration(10))
	}
}

type JsonState struct {
	JsonServiceState string `json:"status"`
}

func checkJsonService(
	checkEndpoint string,
	checkSuccessString string,
	checkRateSeconds int,
	checkName string,
	checkTimeout int) {
	tr := &http.Transport{
		IdleConnTimeout: 1000 * time.Millisecond * time.Duration(checkTimeout),
	}
	client := &http.Client{Transport:tr}
	var returnState int
	for {
		time.Sleep(1000 * time.Millisecond * time.Duration(checkRateSeconds))
		resp, err := client.Get(checkEndpoint)
		if (err == nil) && (resp.StatusCode == 200) {
			if (err != nil){
				fmt.Println(err)
			}
			body, err := ioutil.ReadAll(resp.Body)
			if (err != nil){
				fmt.Println("Unable to read json answer.", err)
				returnState = 0
				Ss[checkName + "_rest_state_up"] = returnState
			} else {
				s, err := getJsonState([]byte(body))
				if (err == nil) && (strings.Contains(s.JsonServiceState, checkSuccessString)) {
					returnState = 1
					Ss[checkName+"_rest_state_up"] = returnState
				} else {
					returnState = 0
					Ss[checkName+"_rest_state_up"] = returnState
				}
			}
		} else {
			returnState = 0
			Ss[checkName + "_rest_state_up"] = returnState
		}
	}
}

func getJsonState (body []byte) (*JsonState, error) {
	var jsonState = new(JsonState)
	err := json.Unmarshal(body, &jsonState)
	if (err != nil){
		fmt.Println("Unable to read json answer.", err)
	}
	return jsonState, err
}

func checkHttpService(
	checkEndpoint string,
	checkSuccessString string,
	checkRateSeconds int,
	checkName string,
	checkTimeout int) {
	var returnState int
	checkSuccessInt, err := strconv.Atoi(checkSuccessString)
	if (err != nil) {
		fmt.Println("Unable to convert string to int success state for endpoint: " + checkEndpoint)
	}
	tr := &http.Transport{
		IdleConnTimeout: 1000 * time.Millisecond * time.Duration(checkTimeout),
		}
	client := &http.Client{Transport:tr}
	for {
		time.Sleep(1000 * time.Millisecond * time.Duration(checkRateSeconds))
		resp, err := client.Get(checkEndpoint)
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