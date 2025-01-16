package pluginmanager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dop251/goja"
	"gopkg.in/yaml.v2"
)

type Plugin struct {
	Name        string
	Author      string
	Description string
	Path        string
	Runtime     *goja.Runtime
}

type PluginManager struct {
	Plugins map[string]*Plugin
}

func NewPluginManager() *PluginManager {
	return &PluginManager{
		Plugins: make(map[string]*Plugin),
	}
}

func (pm *PluginManager) LoadPlugins(pluginDir string) error {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			pluginPath := filepath.Join(pluginDir, entry.Name())
			yamlPath := filepath.Join(pluginPath, "plugin.yaml")
			if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
				fmt.Printf("Skipping %s: plugin.yaml not found\n", entry.Name())
				continue
			}

			plugin, err := pm.loadPlugin(yamlPath, pluginPath)
			if err != nil {
				fmt.Printf("Failed to load plugin %s: %v\n", entry.Name(), err)
				continue
			}

			pm.Plugins[plugin.Name] = plugin
			fmt.Printf("Loaded plugin: %s\n", plugin.Name)
		}
	}

	return nil
}

func (pm *PluginManager) loadPlugin(yamlPath, pluginPath string) (*Plugin, error) {
	// Читаем информацию о плагине из YAML
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.yaml: %w", err)
	}

	var pluginInfo struct {
		PluginName        string `yaml:"PluginName"`
		PluginAuthor      string `yaml:"PluginAuthor"`
		PluginDescription string `yaml:"PluginDescription"`
	}
	if err := yaml.Unmarshal(data, &pluginInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Создаем плагин
	plugin := &Plugin{
		Name:        pluginInfo.PluginName,
		Author:      pluginInfo.PluginAuthor,
		Description: pluginInfo.PluginDescription,
		Path:        pluginPath,
		Runtime:     goja.New(),
	}

	console := map[string]interface{}{
		"log": func(call goja.FunctionCall) goja.Value {
			// Получаем все аргументы
			args := call.Arguments
			var output string
			for _, arg := range args {
				output += fmt.Sprintf("%v ", arg)
			}
			fmt.Println(output) // Печатаем в стандартный вывод Go
			return nil
		},
	}
	plugin.Runtime.Set("console", console)

	// Загружаем другие файлы плагина.
	err = filepath.Walk(pluginPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".js" && filepath.Base(path) != "plugin.yaml" {
			scriptData, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read script %s: %w", path, err)
			}

			_, err = plugin.Runtime.RunString(string(scriptData))
			if err != nil {
				return fmt.Errorf("failed to execute script %s: %w", path, err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return plugin, nil
}

func (pm *PluginManager) GetPlugin(name string) (*Plugin, error) {
	plugin, exists := pm.Plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	return plugin, nil
}
