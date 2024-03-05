package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"false.kr/WebChecker-Node/dto"
)

var Config *dto.ConfigDTO

func Init() {
	Config = configFileScan("config.json")
	if Config == nil {
		os.Exit(3)
	}
}

func configFileCreate(confFile string) bool {
	configData := dto.ConfigDTO{
		Port:       "5200",
		Screenshot: "screenshot/",
	}

	data, err := json.MarshalIndent(&configData, "", "	")
	if err != nil {
		fmt.Println(confFile + " file create failed")
		fmt.Println(err)
		return false
	}

	err2 := os.WriteFile(confFile, data, 0644)
	if err2 != nil {
		fmt.Println(confFile + " file create failed")
		fmt.Println(err)
		return false
	}

	fmt.Println(confFile + " file create")
	return true
}

func configFileScan(confFile string) *dto.ConfigDTO {
	filename, _ := filepath.Abs(confFile)
	jsonFile, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "The system cannot find the file specified.") || strings.Contains(err.Error(), "no such file or directory") {
			if !configFileCreate(confFile) {
				return nil
			}

			return configFileScan(confFile)
		} else {
			return nil
		}
	}

	configData := &dto.ConfigDTO{}
	err = json.Unmarshal([]byte(jsonFile), &configData)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	if _, err := os.Stat(configData.Screenshot); os.IsNotExist(err) {
		if err := os.Mkdir(configData.Screenshot, 0755); err != nil {
			return nil
		}
	}

	return configData
}
