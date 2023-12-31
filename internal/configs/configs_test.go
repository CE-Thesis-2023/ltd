package configs

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

type ConfigsTest func()

func SetupConfigsTest(contents []byte, filePath string, test ConfigsTest) {
	os.Setenv("CONFIG_FILE_PATH", filePath)

	if err := os.WriteFile(filePath, contents, 0644); err != nil {
		log.Fatalf("SetupConfigsTest: create err = %s", err)
		return
	}

	defer func ()  {
		if err := os.Remove(filePath); err != nil {
			log.Fatalf("SetupConfigsTest: remove err = %s", err)
		}
	}()

	test()
}

func TestConfigs_ReadJsonConfigOk(t *testing.T) {
	configs := map[string]interface{}{
		"http": map[string]interface{}{
			"port": 8080,
			"tls": map[string]interface{}{
				"cert": "./certs/cert.pem",
				"key":  "./certs/key.pem",
			},
		},
	}
	contents, _ := json.Marshal(configs)
	SetupConfigsTest(contents, "./configs.json", ConfigsTest(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		Init(ctx)

		result := Get()

		if result == nil {
			t.Fatalf("configs parsed as nil")
		}

		resultContent, _ := json.Marshal(result)
		if string(contents) != string(resultContent) {
			t.Fatalf("configs read is missing")
		}
	}))
}

func TestConfigs_ReadYamlConfigOk(t *testing.T) {
	configs := map[string]interface{}{
		"http": map[string]interface{}{
			"port": 8080,
			"tls": map[string]interface{}{
				"cert": "./certs/cert.pem",
				"key":  "./certs/key.pem",
			},
		},
	}
	contents, _ := yaml.Marshal(configs)
	SetupConfigsTest(contents, "./configs.yaml", ConfigsTest(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		defer cancel()

		Init(ctx)

		result := Get()

		if result == nil {
			t.Fatalf("configs parsed as nil")
		}

		resultContent, _ := yaml.Marshal(result)
		if string(contents) != string(resultContent) {
			t.Fatalf("configs read is missing")
		}
	}))
}
