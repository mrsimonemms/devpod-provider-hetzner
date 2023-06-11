package options

import (
	"fmt"
	"os"
)

type Options struct {
	Token string
}

func FromEnv(skipMachine bool) (*Options, error) {
	retOptions := &Options{}

	var err error

	retOptions.Token, err = fromEnvOrError("TOKEN")
	if err != nil {
		return nil, err
	}

	return retOptions, nil
}

func fromEnvOrError(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf("couldn't find option %s in environment, please make sure %s is defined", name, name)
	}

	return val, nil
}
