package pkg

import (
	"os"
	"testing"
)

func TestLoadYamlConfig(t *testing.T) {
	yamlFile, err := os.Open("../config/configs.yaml")
	if err != nil {
		t.Errorf("error opening the yaml config file >> %v", err)
		return
	}
	config, err := LoadYamlConfigs(yamlFile)
	if err != nil {
		t.Errorf("error reading the yaml configs >> %v", err)
		return
	}
	t.Logf("the strategy is >> %v", config.Strategy)
	for _, service := range config.Services {
		t.Logf("service name >> %v", service.Name)
		for _, replica := range service.Replicas {
			t.Logf("the replica is hosted on >> %v", replica)
		}
	}
}
