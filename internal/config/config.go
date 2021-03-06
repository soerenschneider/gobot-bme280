package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	BotName                = "gobot_bme280"
	defaultLogSensor       = false
	defaultIntervalSeconds = 30
	defaultMetricConfig    = ":9192"
)

type Config struct {
	Placement    string `json:"placement,omitempty"`
	MetricConfig string `json:"metrics_addr,omitempty"`
	IntervalSecs int    `json:"interval_s,omitempty"`
	LogSensor    bool   `json:"log_sensor,omitempty"`
	DisableMqtt  bool   `json:"disable_mqtt"`
	MqttConfig
	SensorConfig
}

func DefaultConfig() Config {
	return Config{
		LogSensor:    defaultLogSensor,
		IntervalSecs: defaultIntervalSeconds,
		MetricConfig: defaultMetricConfig,
		SensorConfig: defaultSensorConfig(),
	}
}

func ConfigFromEnv() Config {
	conf := DefaultConfig()

	placement, err := fromEnv("placement")
	if err == nil {
		conf.Placement = placement
	}

	logSensor, err := fromEnvBool("LOG_SENSOR")
	if err == nil {
		conf.LogSensor = logSensor
	}

	intervalSeconds, err := fromEnvInt("INTERVAL_S")
	if err == nil {
		conf.IntervalSecs = intervalSeconds
	}

	disableMqtt, err := fromEnvBool("DISABLE_MQTT")
	if err == nil {
		conf.DisableMqtt = disableMqtt
	}

	mqttHost, err := fromEnv("MQTT_HOST")
	if err == nil {
		conf.Host = mqttHost
	}

	mqttTopic, err := fromEnv("MQTT_TOPIC")
	if err == nil {
		conf.Topic = mqttTopic
	}

	metricConfig, err := fromEnv("METRICS_ADDR")
	if err == nil {
		conf.MetricConfig = metricConfig
	}

	clientKeyFile, err := fromEnv("SSL_CLIENT_KEY_FILE")
	if err == nil {
		conf.ClientKeyFile = clientKeyFile
	}

	clientCertFile, err := fromEnv("SSL_CLIENT_CERT_FILE")
	if err == nil {
		conf.ClientCertFile = clientCertFile
	}

	conf.SensorConfig.ConfigFromEnv()
	return conf
}

func ReadJsonConfig(filePath string) (*Config, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read config from file: %v", err)
	}

	ret := DefaultConfig()
	err = json.Unmarshal(fileContent, &ret)
	return &ret, err
}

func (conf *Config) Validate() error {
	if conf.Placement == "" {
		return errors.New("empty placement provided")
	}

	if conf.IntervalSecs < 30 {
		return fmt.Errorf("invalid interval: must not be lower than 30 but is %d", conf.IntervalSecs)
	}

	if conf.IntervalSecs > 300 {
		return fmt.Errorf("invalid interval: mut not be greater than 300 but is %d", conf.IntervalSecs)
	}

	if err := conf.SensorConfig.Validate(); err != nil {
		return err
	}

	if conf.DisableMqtt {
		return nil
	}

	return conf.MqttConfig.Validate()
}

func (conf *Config) Print() {
	log.Println("-----------------")
	log.Println("Configuration:")
	log.Printf("Placement=%s", conf.Placement)
	log.Printf("LogSensor=%t", conf.LogSensor)
	log.Printf("MetricConfig=%s", conf.MetricConfig)
	log.Printf("IntervalSecs=%d", conf.IntervalSecs)
	log.Printf("DisableMqtt=%t", conf.DisableMqtt)

	conf.SensorConfig.Print()
	if !conf.DisableMqtt {
		conf.MqttConfig.Print()
	}

	log.Println("-----------------")
}

func computeEnvName(name string) string {
	return fmt.Sprintf("%s_%s", strings.ToUpper(BotName), strings.ToUpper(name))
}

func fromEnv(name string) (string, error) {
	name = computeEnvName(name)
	val := os.Getenv(name)
	if val == "" {
		return "", errors.New("not defined")
	}
	return val, nil
}

func fromEnvInt(name string) (int, error) {
	val, err := fromEnv(name)
	if err != nil {
		return -1, err
	}

	parsed, err := strconv.Atoi(val)
	if err != nil {
		return -1, err
	}
	return parsed, nil
}

func fromEnvBool(name string) (bool, error) {
	val, err := fromEnv(name)
	if err != nil {
		return false, err
	}

	parsed, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}
	return parsed, nil
}
