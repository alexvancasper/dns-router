package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-yaml/yaml"
)

type Config struct {
	Nameservers       []string      `yaml:"nameservers"`
	CorpNameservers   []string      `yaml:"corpnameservers"`
	Blocklist         []string      `yaml:"blocklist"`
	BlockAddress4     string        `yaml:"blockAddress4"`
	BlockAddress6     string        `yaml:"blockAddress6"`
	CorpDomain        string        `yaml:"corpdomain"`
	ExcludeCorpDomain string        `yaml:"excludecorpdomain"`
	ConfigUpdate      bool          `yaml:"configUpdate"`
	UpdateInterval    time.Duration `yaml:"updateInterval"`
}

const (
	COLD bool = true
	WARM bool = false
)

func loadConfig(coldStart bool) (*Config, error) {
	config := &Config{}

	if _, err := os.Stat(*configFile); err != nil {
		return nil, fmt.Errorf("[loadConfig] error: %v", err.Error())
	}

	data, err := os.ReadFile(*configFile)
	if err != nil {
		return nil, fmt.Errorf("[loadConfig] error: %v", err.Error())
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("[loadConfig] error: %v", err.Error())
	}

	if coldStart && config.ConfigUpdate {
		go configWatcher()
	}
	return config, nil
}

func configWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Add(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("Config file updated, reload config")
				c, err := loadConfig(WARM)
				if err != nil {
					log.Println("Bad config: ", err)
				} else {
					log.Println("Config successfully updated")
					config = c
					if !c.ConfigUpdate {
						return
					}
				}
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
