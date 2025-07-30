/*
 * Copyright 2023 steadybit GmbH. All rights reserved.
 */

package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
)

// Specification is the configuration specification for the extension. Configuration values can be applied
// through environment variables. Learn more through the documentation of the envconfig package.
// https://github.com/kelseyhightower/envconfig
type Specification struct {
	DiscoveryAttributesExcludesInstance []string `json:"discoveryAttributesExcludesInstance" split_words:"true" required:"false"`
	InsecureSkipVerify                  bool     `json:"insecureSkipVerify" split_words:"true" default:"false"`
	FetchDelayMillis                    int      `json:"fetchDelayMillis" split_words:"true" default:"0"`
}

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
	if Config.FetchDelayMillis != 0 {
		log.Info().Int("delay", Config.FetchDelayMillis).Msg("Configuration specifies a fetch delay, which will be applied to all Prometheus queries.")
	}
}

func ValidateConfiguration() {
	// You may optionally validate the configuration here.
}
