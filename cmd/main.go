package main

import (
	"fmt"
	"log"
	pluginmanager "py-plugin-test/internal/jsPlugins"
	"sync"
)

func main() {
	manager := pluginmanager.NewPluginManager()

	err := manager.LoadPlugins("./plugins")
	if err != nil {
		log.Fatalf("Failed to load plugins: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	for name, plugin := range manager.Plugins {
		log.Printf("Plugin loaded: %s by %s - %s", name, plugin.Author, plugin.Description)
		go func(name string) {
			defer wg.Done()

			plugin, err := manager.GetPlugin(name)
			if err != nil {
				log.Fatalf("Failed to get plugin: %v", err)
			}

			result, err := plugin.Runtime.RunString(`main()`)
			if err != nil {
				log.Fatalf("Failed to execute plugin function: %v", err)
			}

			log.Printf("Plugin function result: %v", result)
		}(name)
	}

	wg.Wait()
	fmt.Println("All plugins executed")
}
