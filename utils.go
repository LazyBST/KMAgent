package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_FILE_PATH          = "configFilePath"
	CONFIG_SERVICE_ORIGIN_URL = "configServiceOriginUrl"
)

func checkMapEquality(oldMp, newMp map[string]interface{}) bool {
	if len(oldMp) != len(newMp) {
		return false
	}

	for key, oldVal := range oldMp {
		newVal, exists := newMp[key]
		if !exists {
			return false
		}

		if !checkValuesEquality(oldVal, newVal) {
			return false
		}
	}

	return true
}

func checkValuesEquality(oldVal, newVal interface{}) bool {
	switch oldVal := oldVal.(type) {
	case map[string]interface{}:
		newVal, ok := newVal.(map[string]interface{})
		if !ok {
			return false
		}
		return checkMapEquality(oldVal, newVal)
	case []interface{}:
		newVal, ok := newVal.([]interface{})
		if !ok {
			return false
		}
		return checkSliceEquality(oldVal, newVal)
	default:
		return reflect.DeepEqual(oldVal, newVal)
	}
}

func checkSliceEquality(oldVal, newVal interface{}) bool {
	oldSlice := oldVal.([]interface{})
	newSlice := newVal.([]interface{})
	if len(oldSlice) != len(newSlice) {
		return false
	}

	for i := range oldSlice {
		if !checkValuesEquality(oldSlice[i], newSlice[i]) {
			return false
		}
	}

	return true
}

func getEnvFlags() (string, string) {
	yamlFilePath := flag.String(CONFIG_FILE_PATH, "", "Path to the YAML file")
	configServiceOriginUrl := flag.String(CONFIG_SERVICE_ORIGIN_URL, "", "URL of the config service")

	flag.Parse()

	if *yamlFilePath == "" {
		log.Fatalf("flag %s is not set", CONFIG_FILE_PATH)
	}

	if *configServiceOriginUrl == "" {
		log.Fatalf("flag %s is not set", CONFIG_SERVICE_ORIGIN_URL)
	}

	return *yamlFilePath, *configServiceOriginUrl
}

func jsonToYaml(jsonData []byte) ([]byte, error) {
	var jsonObj interface{}
	err := json.Unmarshal(jsonData, &jsonObj)
	if err != nil {
		return nil, err
	}

	yamlData, err := yaml.Marshal(jsonObj)
	if err != nil {
		return nil, err
	}

	return yamlData, nil
}

func saveYamlToFile(yamlData []byte, filename string) error {
	err := os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func isConfigExists(configPath string) bool {
	if _, err := os.Stat(configPath); err != nil {
		return false
	}

	return true
}
