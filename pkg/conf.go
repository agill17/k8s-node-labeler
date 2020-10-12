package pkg

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)
// TODO: take input to exclude nodes with k:v as labels ( excludeNodesWithLabels )
type desiredConf struct {
	DesiredLabels map[string]string `yaml:"desiredLabels"`
}

func NewDesiredConf(confFile string) (*desiredConf, error) {
	rawConf, errReading := ioutil.ReadFile(confFile)
	if errReading != nil {
		return nil, errReading
	}
	config := &desiredConf{}
	err := yaml.Unmarshal(rawConf,config)
	return config, err
}