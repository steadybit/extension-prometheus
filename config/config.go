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
	InsecureSkipVerify                  bool     `json:"insecureSkipVerify" split_words:"true" default:"false" required:"false"`
	EnableRequestLogging                bool     `json:"enableRequestLogging" split_words:"true" default:"false" required:"false"`
	AdditionalRequestParams             []string `json:"additionalRequestParams" split_words:"true" required:"false"`
}

var (
	Config Specification
)

func ParseConfiguration() {
	err := envconfig.Process("steadybit_extension", &Config)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse configuration from environment.")
	}
	if len(Config.AdditionalRequestParams)%2 != 0 {
		log.Fatal().Msgf("Additional request parameters must be provided in key-value pairs, but an odd number of parameters was provided.")
	}
}

func ValidateConfiguration() {
	// You may optionally validate the configuration here.
}
