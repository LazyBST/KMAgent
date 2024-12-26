package utils

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_FILE_PATH_FLAG          = "configFilePath"
	CONFIG_SERVICE_ORIGIN_URL_FLAG = "configServiceOriginUrl"
	RUN_AS_SERVICE_FLAG            = "run-as-service"
)

const SERVICE_NAME = "kmagent"

func CheckMapEquality(oldMp, newMp map[string]interface{}) bool {
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
		return CheckMapEquality(oldVal, newVal)
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

func GetEnvFlags() (string, string, bool) {
	yamlFilePath := flag.String(CONFIG_FILE_PATH_FLAG, "", "Path to the YAML file")
	configServiceOriginUrl := flag.String(CONFIG_SERVICE_ORIGIN_URL_FLAG, "", "URL of the config service")
	runAsService := flag.Bool("run-as-service", false, "Run as a service")

	flag.Parse()

	if *yamlFilePath == "" {
		log.Fatalf("flag %s is not set", CONFIG_FILE_PATH_FLAG)
	}

	if *configServiceOriginUrl == "" {
		log.Fatalf("flag %s is not set", CONFIG_SERVICE_ORIGIN_URL_FLAG)
	}

	return *yamlFilePath, *configServiceOriginUrl, *runAsService
}

func JsonToYaml(jsonData []byte) ([]byte, error) {
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

func SaveYamlToFile(yamlData []byte, filename string) error {
	err := os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func IsConfigExists(configPath string) bool {
	if _, err := os.Stat(configPath); err != nil {
		return false
	}

	return true
}

func getExecPath() string {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Error getting executable path: %v", err)
	}

	return execPath
}

func BoolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
