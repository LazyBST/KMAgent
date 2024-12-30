package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/lazybst/kmagent/pkg/utils"
	"gopkg.in/yaml.v2"
)

type agent struct{}

func (p *agent) Start(s service.Service) error {
	if service.Interactive() {
		log.Print("manage")
		return p.manage(s)
	}
	log.Print("start")
	go p.run()
	return nil
}

func (p *agent) manage(s service.Service) error {
	if status, err := s.Status(); err == service.ErrNotInstalled {
		log.Printf("installing...")
		if err := s.Install(); err != nil {
			return err
		}
		log.Printf("installed")
	} else if err != nil {
		return err
	} else if status == service.StatusUnknown {
		log.Printf("service unknown")
		if err := s.Uninstall(); err == nil {
			log.Printf("uninstalled")
		}
	} else if status == service.StatusStopped {
		log.Printf("service stopped. starting...")
		if err := s.Start(); err != nil {
			return err
		}
		log.Printf("started")
	} else {
		log.Printf("service running")
		if err := s.Uninstall(); err == nil {
			log.Print("uninstalled")
		}
	}
	return nil
}

func (p *agent) run() {
	log.Println("Starting agent...")
	initAgent()
}
func (p *agent) Stop(s service.Service) error {
	log.Print("stop")
	stopCollector()
	time.Sleep(100 * time.Millisecond)
	return nil
}

func initAgent() {
	log.Println("Initializing agent...")
	isNewConfigCh := make(chan bool, 1)

	go pollForConfig(CONFIG_SVC_ORIGIN+"/config", isNewConfigCh)
	go pollStatus(CONFIG_SVC_ORIGIN + "/status")

	checkForConfigAndStartCollector(isNewConfigCh)
}

func checkForConfigAndStartCollector(isNewConfigCh <-chan bool) {
	if utils.IsConfigExists(CONFIG_PATH) && collector == nil {
		startCollector()
		sendStatusUpdate(CONFIG_SVC_ORIGIN)
	}

	for <-isNewConfigCh {
		log.Println("restarting collector...")
		startCollector()
		sendStatusUpdate(CONFIG_SVC_ORIGIN)
	}
}

func pollForConfig(configEndpoint string, isNewConfigCh chan<- bool) {
	log.Println("Initializing config polling...")
	for {
		resp, err := http.Get(configEndpoint)

		if err != nil {
			log.Printf("Error getting config from %s: %v", CONFIG_SVC_ORIGIN, err)
			time.Sleep(10 * time.Second)
			continue
		}

		jsonData, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Fatalf("Error reading response body: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		yamlData, err := utils.JsonToYaml(jsonData)
		if err != nil {
			log.Println("Error converting JSON to YAML:", err)
			time.Sleep(10 * time.Second)
			continue
		}

		// if no new config detected continue
		if !isDiffConfig(yamlData, CONFIG_PATH) {
			time.Sleep(10 * time.Second)
			continue
		}

		log.Println("new config detected")

		err = utils.SaveYamlToFile(yamlData, CONFIG_PATH)
		if err != nil {
			log.Println("Error saving config file:", err)
			continue
		}

		isNewConfigCh <- true

		time.Sleep(10 * time.Second)
	}
}

func isDiffConfig(newConfig []byte, prevConfigFile string) bool {
	oldConfig, err := os.ReadFile(prevConfigFile)
	if err != nil {
		log.Println("no config file found:", err)
		return true
	}

	oldConfigMp := make(map[string]interface{})
	newConfigMp := make(map[string]interface{})

	if err = yaml.Unmarshal(oldConfig, &oldConfigMp); err != nil {
		log.Println("Error unmarshalling old config:", err)
		return false
	}
	if err = yaml.Unmarshal(newConfig, &newConfigMp); err != nil {
		log.Println("Error unmarshalling new config:", err)
		return false
	}

	return !utils.CheckMapEquality(oldConfigMp, newConfigMp)
}

func pollStatus(url string) {
	log.Println("Initializing status polling...")
	for {
		sendStatusUpdate(url)
		time.Sleep(10 * time.Second)
	}
}

func sendStatusUpdate(url string) {
	if collector == nil {
		return
	}
	payload := make(map[string]interface{})
	payload["status"] = collector.GetState().String()

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		log.Println("Error marshalling payload:", err)
		return
	}

	if _, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload)); err != nil {
		log.Println("Error posting status:", err)
	}
}
