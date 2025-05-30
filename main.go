package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v3"

	faker "github.com/bxcodec/faker/v3"
	"github.com/spf13/cobra"
)

type Endpoint struct {
	Path string `yaml:"path" json:"path"`
	Method string `yaml:"method" json:"method"`
	Data string `yaml:"data" json:"data"`
	Count int `yaml:"count" json:"count"`
}

type Config struct {
	Port int `yaml:"port" json:"port"`
	Endpoint []Endpoint `yaml:"endpoints" json:"endpoints"`
}

func loadConfig(path string) (*Config, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	if path[len(path)-5:] == ".yaml" || path[len(path)-4:] == ".yml" {
		err = yaml.Unmarshal(file, config)
	} else {
		err = json.Unmarshal(file, config)
	}
	return config, err 
}

func generateFakeData(schema string, count int) ([]map[string]interface{}, error) {
	var fake []map[string]interface{}
	for i:= 0; i < count; i++ {
		data := map[string]interface{}{}
		err := faker.FakeData(&data)
		if err != nil {
			return nil, err 
		}
		fake = append(fake, data)
	}
	return fake, nil
}

func startServer(config *Config) {
	for _, ep := range config.Endpoint {
		path := ep.Path
		method := ep.Method
		dataCount := ep.Count
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
				return
			}
			data, err := generateFakeData(ep.Data, dataCount)
			if err != nil {
				http.Error(w, "Failed to generate data", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(data)
		}) 
	}
	log.Printf("Starting mock server on :%d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",config.Port),nil))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var configPath string

	var rootCmd = &cobra.Command{
		Use: "mock-api-server",
		Short: "Lightweight TUI/mock REST API server",
		Run: func(cmd *cobra.Command, args []string) {
			config, err := loadConfig(configPath)
			if err != nil {
				log.Fatalf("Failed to load config: %v", err)
			}
			startServer(config)
		},
	}

	rootCmd.Flags().StringVarP(&configPath, "config", "c", "mock.yaml", "Path to mock config file")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
