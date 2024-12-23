// Code generated by "go.opentelemetry.io/collector/cmd/builder".

// Program kmagent is an OpenTelemetry Collector binary.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
	envprovider "go.opentelemetry.io/collector/confmap/provider/envprovider"
	fileprovider "go.opentelemetry.io/collector/confmap/provider/fileprovider"
	httpprovider "go.opentelemetry.io/collector/confmap/provider/httpprovider"
	httpsprovider "go.opentelemetry.io/collector/confmap/provider/httpsprovider"
	yamlprovider "go.opentelemetry.io/collector/confmap/provider/yamlprovider"
	"go.opentelemetry.io/collector/otelcol"
	"gopkg.in/yaml.v2"
)

var (
	CONFIG_PATH       = "/tmp/otelcol.yaml"
	CONFIG_SVC_ORIGIN = "http://localhost:3000"
)

var collector *otelcol.Collector

func main() {
	CONFIG_PATH, CONFIG_SVC_ORIGIN = getEnvFlags()

	isNewConfigCh := make(chan bool, 1)

	go listenForInterrupt()

	go pollForConfig(CONFIG_SVC_ORIGIN+"/config", isNewConfigCh)
	go pollStatus(CONFIG_SVC_ORIGIN + "/status")

	checkForConfigAndStartCollector(isNewConfigCh)
}

func listenForInterrupt() {
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt)

	go func(interruptCh <-chan os.Signal) {
		for sig := range interruptCh {
			if sig == os.Interrupt && collector != nil {
				log.Println("Interrupt signal received, shutting down...")
				collector.Shutdown()
				collector = nil
				time.Sleep(1 * time.Second)
				os.Exit(0)
			} else {
				os.Exit(0)
			}
		}
	}(interruptCh)
}

func checkForConfigAndStartCollector(isNewConfigCh <-chan bool) {
	if isConfigExists(CONFIG_PATH) && collector == nil {
		startCollector()
		sendStatusUpdate(CONFIG_SVC_ORIGIN)
	}

	for <-isNewConfigCh {
		log.Println("restarting collector...")
		startCollector()
		sendStatusUpdate(CONFIG_SVC_ORIGIN)
	}
}

func runInteractive(params otelcol.CollectorSettings) {
	go func(params otelcol.CollectorSettings) {
		var err error
		collector, err = otelcol.NewCollector(params)

		if err != nil {
			log.Fatalf("collector instance creation failed with error: %v", err)
		}

		err = collector.Run(context.Background())

		if err != nil {
			log.Fatalf("collector server run finished with error: %v", err)
		}
	}(params)
}

func pollForConfig(configEndpoint string, isNewConfigCh chan<- bool) {
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

		yamlData, err := jsonToYaml(jsonData)
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

		err = saveYamlToFile(yamlData, CONFIG_PATH)
		if err != nil {
			log.Println("Error saving config file:", err)
			continue
		}

		isNewConfigCh <- true

		time.Sleep(10 * time.Second)
	}
}

func startCollector() {
	if collector != nil {
		collector.Shutdown()
		collector = nil
		log.Println("old collector instance stopped")
	}

	info := component.BuildInfo{
		Command:     "kmagent",
		Description: "Agent-as-collector POC",
		Version:     "0.0.1",
	}

	set := otelcol.CollectorSettings{
		BuildInfo: info,
		Factories: components,
		ConfigProviderSettings: otelcol.ConfigProviderSettings{
			ResolverSettings: confmap.ResolverSettings{
				URIs: []string{CONFIG_PATH},
				ProviderFactories: []confmap.ProviderFactory{
					envprovider.NewFactory(),
					fileprovider.NewFactory(),
					httpprovider.NewFactory(),
					httpsprovider.NewFactory(),
					yamlprovider.NewFactory(),
				},
			},
		},
	}

	runInteractive(set)
	log.Println("collector started")
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

	return !checkMapEquality(oldConfigMp, newConfigMp)
}

func pollStatus(url string) {
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