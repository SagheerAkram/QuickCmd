package plugins

import (
	"fmt"
	"sync"
)

// Registry manages plugin registration and lifecycle
type Registry struct {
	mu      sync.RWMutex
	plugins map[string]Plugin
	metadata map[string]*PluginMetadata
	hooks   map[HookType][]HookFunc
}

// NewRegistry creates a new plugin registry
func NewRegistry() *Registry {
	return &Registry{
		plugins:  make(map[string]Plugin),
		metadata: make(map[string]*PluginMetadata),
		hooks:    make(map[HookType][]HookFunc),
	}
}

// Register registers a plugin with the registry
func (r *Registry) Register(plugin Plugin, metadata *PluginMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}
	
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s already registered", name)
	}
	
	r.plugins[name] = plugin
	r.metadata[name] = metadata
	
	return nil
}

// Unregister removes a plugin from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	
	delete(r.plugins, name)
	delete(r.metadata, name)
	
	return nil
}

// Get retrieves a plugin by name
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	
	return plugin, nil
}

// List returns all registered plugins
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugins := make([]Plugin, 0, len(r.plugins))
	for _, plugin := range r.plugins {
		plugins = append(plugins, plugin)
	}
	
	return plugins
}

// ListEnabled returns all enabled plugins
func (r *Registry) ListEnabled() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	plugins := make([]Plugin, 0, len(r.plugins))
	for name, plugin := range r.plugins {
		if meta, exists := r.metadata[name]; exists && meta.Enabled {
			plugins = append(plugins, plugin)
		}
	}
	
	return plugins
}

// GetMetadata retrieves plugin metadata
func (r *Registry) GetMetadata(name string) (*PluginMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	metadata, exists := r.metadata[name]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", name)
	}
	
	return metadata, nil
}

// Enable enables a plugin
func (r *Registry) Enable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	metadata, exists := r.metadata[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	
	metadata.Enabled = true
	return nil
}

// Disable disables a plugin
func (r *Registry) Disable(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	metadata, exists := r.metadata[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}
	
	metadata.Enabled = false
	return nil
}

// RegisterHook registers a hook function
func (r *Registry) RegisterHook(hookType HookType, fn HookFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.hooks[hookType] = append(r.hooks[hookType], fn)
}

// ExecuteHooks executes all hooks of a given type
func (r *Registry) ExecuteHooks(hookType HookType, ctx Context, data interface{}) error {
	r.mu.RLock()
	hooks := r.hooks[hookType]
	r.mu.RUnlock()
	
	for _, hook := range hooks {
		if err := hook(ctx, data); err != nil {
			return err
		}
	}
	
	return nil
}

// Global registry instance
var globalRegistry = NewRegistry()

// DefaultRegistry returns the global plugin registry
func DefaultRegistry() *Registry {
	return globalRegistry
}

// Register registers a plugin with the default registry
func Register(plugin Plugin, metadata *PluginMetadata) error {
	return globalRegistry.Register(plugin, metadata)
}

// Get retrieves a plugin from the default registry
func Get(name string) (Plugin, error) {
	return globalRegistry.Get(name)
}

// List returns all plugins from the default registry
func List() []Plugin {
	return globalRegistry.List()
}

// ListEnabled returns all enabled plugins from the default registry
func ListEnabled() []Plugin {
	return globalRegistry.ListEnabled()
}
