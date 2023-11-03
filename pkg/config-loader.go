package pkg

import (
	"io"
	"log"

	"github.com/FadyGamilM/gobalancer/config"
	"gopkg.in/yaml.v3"
)

func LoadYamlConfigs(reader io.Reader) (*config.BalancerConfig, error) {
	buf, err := io.ReadAll(reader)
	if err == io.EOF {
		log.Println("end of configs yaml file")
	} else if err != nil {
		return nil, err
	}
	conf := &config.BalancerConfig{}
	// convert data from yaml to config struct
	if err := yaml.Unmarshal(buf, conf); err != nil {
		return nil, err
	}
	return conf, nil
}
