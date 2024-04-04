package configs

import (
	"context"
	"encoding/json"
	custerror "github.com/CE-Thesis-2023/ltd/src/internal/error"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var globalConfigs *Configs

type Configs struct {
	Public           HttpConfigs       `json:"public,omitempty" yaml:"public,omitempty"`
	Private          HttpConfigs       `json:"private,omitempty" yaml:"private,omitempty"`
	Logger           LoggerConfigs     `json:"logger,omitempty" yaml:"logger,omitempty"`
	MqttStore        EventStoreConfigs `json:"mqttStore,omitempty" yaml:"mqttStore,omitempty"`
	CloudMediaServer MediaMtxConfigs   `json:"cloudMediaServer,omitempty" yaml:"cloudMediaServer,omitempty"`
	DeviceInfo       DeviceInfoConfigs `json:"deviceInfo,omitempty" yaml:"deviceInfo,omitempty"`
	Ffmpeg           FfmpegConfigs     `json:"ffmpeg,omitempty" yaml:"ffmpeg,omitempty"`
}

func (c Configs) String() string {
	configBytes, _ := json.Marshal(c)
	return string(configBytes)
}

func Init(ctx context.Context) {
	configs, err := readConfig()
	if err != nil {
		log.Fatal(err)
		return
	}
	globalConfigs = configs
}

func Get() *Configs {
	return globalConfigs
}

type HttpConfigs struct {
	Name string           `json:"name,omitempty" yaml:"name,omitempty"`
	Port int              `json:"port,omitempty" yaml:"port,omitempty"`
	Tls  TlsConfig        `json:"tls,omitempty" yaml:"tls,omitempty"`
	Auth BasicAuthConfigs `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type TlsConfig struct {
	Cert      string `json:"cert,omitempty" yaml:"cert,omitempty"`
	Key       string `json:"key,omitempty" yaml:"key,omitempty"`
	Authority string `json:"authority,omitempty" yaml:"authority,omitempty"`
	Enabled   bool   `json:"enabled,omitempty" yaml:"enabled,omitempty"`
}

func (c TlsConfig) IsEnabled() bool {
	if len(c.Cert) > 0 && len(c.Key) > 0 {
		return true
	}
	if c.Enabled {
		return true
	}
	return false
}

type LoggerConfigs struct {
	Level    string `json:"level,omitempty" yaml:"level,omitempty"`
	Encoding string `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

type BasicAuthConfigs struct {
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Token    string `json:"token,omitempty" yaml:"token,omitempty"`
}

type EventStoreConfigs struct {
	Tls      TlsConfig `json:"tls,omitempty" yaml:"tls,omitempty"`
	Host     string    `json:"host,omitempty" yaml:"host,omitempty"`
	Port     int       `json:"port,omitempty" yaml:"port,omitempty"`
	Name     string    `json:"name,omitempty" yaml:"name,omitempty"`
	Enabled  bool      `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Username string    `json:"username,omitempty" yaml:"username,omitempty"`
	Password string    `json:"password,omitempty" yaml:"password,omitempty"`
	Level    string    `json:"level,omitempty" yaml:"level,omitempty"`
}

type MediaMtxConfigs struct {
	Host         string        `json:"host,omitempty" yaml:"host,omitempty"`
	Port         int           `json:"port,omitempty" yaml:"port,omitempty"`
	Username     string        `json:"username,omitempty" yaml:"username,omitempty"`
	Password     string        `json:"password,omitempty" yaml:"password,omitempty"`
	PublishPorts MediaMtxPorts `json:"publishPorts,omitempty" yaml:"publishPorts,omitempty"`
}

type MediaMtxPorts struct {
	WebRtc int `json:"webRtc,omitempty" yaml:"webRtc,omitempty"`
	Srt    int `json:"srt,omitempty" yaml:"srt,omitempty"`
	Rtmp   int `json:"rtmp,omitempty" yaml:"rtmp,omitempty"`
}

type FfmpegConfigs struct {
	BinaryPath string `json:"binaryPath,omitempty" yaml:"binaryPath,omitempty"`
}

type DeviceInfoConfigs struct {
	DeviceId       string `json:"deviceId,omitempty" yaml:"deviceId,omitempty"`
	Username       string `json:"username,omitempty" yaml:"username,omitempty"`
	Token          string `json:"token,omitempty" yaml:"token,omitempty"`
	CloudApiServer string `json:"cloudApiServer,omitempty" yaml:"cloudApiServer,omitempty"`
}

func (c *EventStoreConfigs) HasAuth() bool {
	return len(c.Username) > 0 && len(c.Password) > 0
}

func readConfig() (*Configs, error) {
	path, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	configFile, err := readConfigFile(path)
	if err != nil {
		return nil, err
	}

	configs, err := parseConfig(configFile)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func getConfigFilePath() (string, error) {
	path := os.Getenv(ENV_CONFIG_FILE_PATH)
	if len(path) == 0 {
		return "", custerror.FormatNotFound("ENV_CONFIG_FILE_PATH not found, unable to read configurations")
	}
	return path, nil
}

func readConfigFile(path string) ([]byte, error) {
	fs, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, custerror.FormatNotFound("readConfigFile: file not found")
		}
		return nil, custerror.FormatInternalError("readConfigFile: err = %s", err)
	}

	contents, err := os.ReadFile(fs.Name())
	if err != nil {
		return nil, custerror.FormatInternalError("readConfigFile: err = %s", err)
	}

	return contents, nil
}

func parseConfig(contents []byte) (*Configs, error) {
	configs := &Configs{}
	if jsonErr := json.Unmarshal(contents, configs); jsonErr != nil {
		if yamlErr := yaml.Unmarshal(contents, configs); yamlErr != nil {
			return nil, custerror.FormatInvalidArgument("parseConfig: config parse JSON err = %s YAML err = %s", jsonErr, yamlErr)
		}
	}
	return configs, nil
}
