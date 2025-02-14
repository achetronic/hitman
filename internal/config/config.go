package config

import (
	"os"

	"hitman/api/v1alpha1"

	"gopkg.in/yaml.v3"
)

// Marshal TODO
func Marshal(config *v1alpha1.ConfigT) (bytes []byte, err error) {
	bytes, err = yaml.Marshal(config)
	return bytes, err
}

// Unmarshal TODO
func Unmarshal(bytes []byte) (*v1alpha1.ConfigT, error) {
	config := &v1alpha1.ConfigT{}
	err := yaml.Unmarshal(bytes, config)
	return config, err
}

// ReadFile TODO
func ReadFile(filepath string) (*v1alpha1.ConfigT, error) {
	var fileBytes []byte
	fileBytes, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	config, err := Unmarshal(fileBytes)
	return config, err
}
