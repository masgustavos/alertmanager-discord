package config

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

// DiscordChannel contains the necessary Discord DiscordChannel properties
// for the application
type DiscordChannel struct {
	Name                        string   `json:"name" yaml:"name"`
	WebhookURL                  string   `json:"webhookURL" yaml:"webhookURL"`
	RolesToMention              []string `json:"rolesToMention" yaml:"rolesToMention"`
	SeveritiesToMention         []string `json:"severitiesToMention" yaml:"severitiesToMention"`
	SeveritiesToIgnoreWhenAlone []string `json:"severitiesToIgnoreWhenAlone" yaml:"severitiesToIgnoreWhenAlone"`
}

// EmbedAppearance defines the Embed's color and Emoji to be used in the title
type EmbedAppearance struct {
	Color int    `json:"color" yaml:"color"`
	Emoji string `json:"emoji" yaml:"emoji"`
}

// Config defines the (.yaml|.json) config structured to be used by the app
type Config struct {
	PrometheusURL               string                     `json:"prometheusURL" yaml:"prometheusURL"`
	AvatarURL                   string                     `json:"avatarURL" yaml:"avatarURL"`
	Username                    string                     `json:"username" yaml:"username"`
	MessageType                 string                     `json:"messageType" yaml:"messageType"`
	Status                      map[string]EmbedAppearance `json:"status" yaml:"status"`
	FiringCountToMention        int                        `json:"firingCountToMention" yaml:"firingCountToMention"`
	RolesToMention              []string                   `json:"rolesToMention" yaml:"rolesToMention"`
	SeveritiesToMention         []string                   `json:"severitiesToMention" yaml:"severitiesToMention"`
	SeveritiesToIgnoreWhenAlone []string                   `json:"severitiesToIgnoreWhenAlone" yaml:"severitiesToIgnoreWhenAlone"`
	Severity                    struct {
		Label  string                     `json:"label" yaml:"label"`
		Values map[string]EmbedAppearance `json:"values" yaml:"values"`
	}
	DiscordChannels map[string]DiscordChannel `json:"channels" yaml:"channels"`
}

// LoadUserConfig provides a Config struct to be used throughout the application
func LoadUserConfig() *Config {

	configFilePath := getEnv("CONFIG_PATH", "./config.default.yaml")
	defaultConfig := loadConfigurationFile("./config.default.yaml")
	userConfig := loadConfigurationFile(configFilePath)

	err := mergo.Merge(&userConfig, defaultConfig)
	if err != nil {
		log.Fatalln(err)
	}

	yamlConfig, err := yaml.Marshal(userConfig)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Using the following config:\n\n=======\n\n%s\n\n========\n\n", string(yamlConfig))

	return &userConfig
}

func loadConfigurationFile(file string) Config {
	var config Config

	configFile, err := os.Open(file)
	if err != nil {
		log.Fatalln(err)
	}

	defer configFile.Close()

	if strings.HasSuffix(file, ".json") {
		jsonParser := json.NewDecoder(configFile)
		err := jsonParser.Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(file, ".yaml") || strings.HasSuffix(file, ".yml") {
		yamlParser := yaml.NewDecoder(configFile)
		err := yamlParser.Decode(&config)
		if err != nil {
			log.Fatal(err)
		}
	}

	return config
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
