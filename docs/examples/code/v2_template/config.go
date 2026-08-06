// discord.go code
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type templateConfig struct {
	Color   string `json:"color"`
	Prefix  string `json:"prefix"`
	BotName string `json:"botName"`
}

func loadTemplateConfig() templateConfig {
	config := templateConfig{Color: "000000", Prefix: "?", BotName: "zar"}
	path := os.Getenv("V2_CONFIG")
	if path == "" {
		path = "config.json"
	}
	paths := []string{path}
	if path == "config.json" {
		paths = append(paths, filepath.Join("docs", "examples", "code", "v2_template", path))
	}
	for _, candidate := range paths {
		if data, err := os.ReadFile(candidate); err == nil {
			_ = json.Unmarshal(data, &config)
			break
		}
	}
	return config
}

func (c templateConfig) accentColor() int {
	value := strings.TrimPrefix(strings.TrimSpace(c.Color), "#")
	color, err := strconv.ParseInt(value, 16, 32)
	if err != nil {
		return 0
	}
	return int(color)
}
