package plugins

import (
	"fmt"
)

// Loader handles plugin loading and initialization
type Loader struct {
	registry *Registry
}

// NewLoader creates a new plugin loader
func NewLoader(registry *Registry) *Loader {
	return &Loader{
		registry: registry,
	}
}

// LoadBuiltins loads all built-in plugins
func (l *Loader) LoadBuiltins() error {
	// Built-in plugins are registered via init() functions
	// This method is a placeholder for future dynamic loading
	return nil
}

// LoadPlugin loads a single plugin
func (l *Loader) LoadPlugin(plugin Plugin, metadata *PluginMetadata) error {
	if metadata == nil {
		metadata = &PluginMetadata{
			Name:    plugin.Name(),
			Version: "1.0.0",
			Enabled: true,
		}
	}
	
	if metadata.Name == "" {
		metadata.Name = plugin.Name()
	}
	
	return l.registry.Register(plugin, metadata)
}

// UnloadPlugin unloads a plugin by name
func (l *Loader) UnloadPlugin(name string) error {
	return l.registry.Unregister(name)
}

// ReloadPlugin reloads a plugin
func (l *Loader) ReloadPlugin(name string, plugin Plugin, metadata *PluginMetadata) error {
	if err := l.UnloadPlugin(name); err != nil {
		// If plugin doesn't exist, just load it
		if err.Error() != fmt.Sprintf("plugin %s not found", name) {
			return err
		}
	}
	
	return l.LoadPlugin(plugin, metadata)
}

// DefaultLoader returns a loader for the default registry
func DefaultLoader() *Loader {
	return NewLoader(DefaultRegistry())
}

// LoadBuiltins loads built-in plugins into the default registry
func LoadBuiltins() error {
	return DefaultLoader().LoadBuiltins()
}
