// Copyright (c) 2016 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package loadtest

import (
	"reflect"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type LoadTestConfig struct {
	LoadtestEnviromentConfig  LoadtestEnviromentConfig
	ConnectionConfiguration   ConnectionConfiguration
	UserEntitiesConfiguration UserEntitiesConfiguration
	ResultsConfiguration      ResultsConfiguration
}

type UserEntitiesConfiguration struct {
	TestLengthMinutes                 int
	NumActiveEntities                 int
	ActionRateMilliseconds            int
	ActionRateMaxVarianceMilliseconds int
	EnableRequestTiming               bool
	ChannelLinkChance                 float64
	UploadImageChance                 float64
	DoStatusPolling                   bool
	RandomizeEntitySelection          bool
}

type ConnectionConfiguration struct {
	ServerURL            string
	WebsocketURL         string
	PProfURL             string
	DriverName           string
	DataSource           string
	DBEndpoint           string // deprecated
	LocalCommands        bool
	SSHHostnamePort      string
	SSHUsername          string
	SSHPassword          string
	SSHKey               string
	MattermostInstallDir string
	ConfigFileLoc        string
	AdminEmail           string
	AdminPassword        string
	SkipBulkload         bool
	WaitForServerStart   bool
}

type ResultsConfiguration struct {
	CustomReportText     string
	SendReportToMMServer bool
	ResultsServerURL     string
	ResultsChannelId     string
	ResultsUsername      string
	ResultsPassword      string
	PProfDelayMinutes    int
	PProfLength          int
}

func GetConfig() (*LoadTestConfig, error) {
	viper.SetConfigName("loadtestconfig")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.SetEnvPrefix("mmloadtest")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var cfg *LoadTestConfig

	if err := unmarshalConfigStruct(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func unmarshalConfigStruct(configStruct interface{}) error {
	return viper.Unmarshal(configStruct)
}

func unmarshalConfigSubStruct(configStruct interface{}) error {
	return viper.Sub(reflect.ValueOf(configStruct).Elem().Type().Name()).Unmarshal(configStruct)
}

func SetIntFlag(flags *pflag.FlagSet, full, short, helpText, configFileSetting string, defaultValue int) {
	flags.IntP(full, short, defaultValue, helpText)
	viper.SetDefault(configFileSetting, defaultValue)
	viper.BindPFlag(configFileSetting, flags.Lookup(full))
}

func SetBoolFlag(flags *pflag.FlagSet, full, short, helpText, configFileSetting string, defaultValue bool) {
	flags.BoolP(full, short, defaultValue, helpText)
	viper.SetDefault(configFileSetting, defaultValue)
	viper.BindPFlag(configFileSetting, flags.Lookup(full))
}
