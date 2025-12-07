package plugins

import (
	"testing"
	"time"
)

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	
	plugin := &mockPlugin{name: "test"}
	metadata := &PluginMetadata{
		Name:    "test",
		Version: "1.0.0",
		Enabled: true,
	}
	
	err := registry.Register(plugin, metadata)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	
	// Try to register again - should fail
	err = registry.Register(plugin, metadata)
	if err == nil {
		t.Error("Register() should fail for duplicate plugin")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	
	plugin := &mockPlugin{name: "test"}
	metadata := &PluginMetadata{Name: "test", Enabled: true}
	
	registry.Register(plugin, metadata)
	
	retrieved, err := registry.Get("test")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	
	if retrieved.Name() != "test" {
		t.Errorf("Get() returned wrong plugin: %s", retrieved.Name())
	}
	
	// Try to get non-existent plugin
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Get() should fail for non-existent plugin")
	}
}

func TestRegistry_List(t *testing.T) {
	registry := NewRegistry()
	
	plugins := []*mockPlugin{
		{name: "plugin1"},
		{name: "plugin2"},
		{name: "plugin3"},
	}
	
	for _, p := range plugins {
		registry.Register(p, &PluginMetadata{Name: p.name, Enabled: true})
	}
	
	list := registry.List()
	if len(list) != 3 {
		t.Errorf("List() returned %d plugins, want 3", len(list))
	}
}

func TestRegistry_EnableDisable(t *testing.T) {
	registry := NewRegistry()
	
	plugin := &mockPlugin{name: "test"}
	metadata := &PluginMetadata{Name: "test", Enabled: true}
	
	registry.Register(plugin, metadata)
	
	// Disable plugin
	err := registry.Disable("test")
	if err != nil {
		t.Fatalf("Disable() error = %v", err)
	}
	
	// Check it's disabled
	enabled := registry.ListEnabled()
	if len(enabled) != 0 {
		t.Error("ListEnabled() should return empty list after disable")
	}
	
	// Enable plugin
	err = registry.Enable("test")
	if err != nil {
		t.Fatalf("Enable() error = %v", err)
	}
	
	// Check it's enabled
	enabled = registry.ListEnabled()
	if len(enabled) != 1 {
		t.Error("ListEnabled() should return 1 plugin after enable")
	}
}

func TestRegistry_Hooks(t *testing.T) {
	registry := NewRegistry()
	
	called := false
	hookFunc := func(ctx Context, data interface{}) error {
		called = true
		return nil
	}
	
	registry.RegisterHook(HookPreTranslate, hookFunc)
	
	ctx := Context{
		WorkingDir: "/test",
		User:       "testuser",
		Timestamp:  time.Now(),
	}
	
	err := registry.ExecuteHooks(HookPreTranslate, ctx, nil)
	if err != nil {
		t.Fatalf("ExecuteHooks() error = %v", err)
	}
	
	if !called {
		t.Error("Hook was not called")
	}
}

func TestTranslateWithPlugins(t *testing.T) {
	// Create a test registry
	registry := NewRegistry()
	
	// Register mock plugin
	plugin := &mockPlugin{
		name: "test",
		translateFunc: func(ctx Context, prompt string) ([]*Candidate, error) {
			return []*Candidate{
				{
					Command:     "test command",
					Explanation: "test explanation",
					Confidence:  90,
					RiskLevel:   RiskSafe,
				},
			}, nil
		},
	}
	
	registry.Register(plugin, &PluginMetadata{Name: "test", Enabled: true})
	
	// Replace global registry temporarily
	oldRegistry := globalRegistry
	globalRegistry = registry
	defer func() { globalRegistry = oldRegistry }()
	
	ctx := Context{
		WorkingDir: "/test",
		User:       "testuser",
		Timestamp:  time.Now(),
	}
	
	coreCandidates := []*Candidate{
		{
			Command:     "core command",
			Explanation: "core explanation",
			Confidence:  85,
			RiskLevel:   RiskSafe,
		},
	}
	
	candidates, err := TranslateWithPlugins(ctx, "test prompt", coreCandidates)
	if err != nil {
		t.Fatalf("TranslateWithPlugins() error = %v", err)
	}
	
	if len(candidates) != 2 {
		t.Errorf("TranslateWithPlugins() returned %d candidates, want 2", len(candidates))
	}
	
	// Check plugin name is set
	if candidates[1].PluginName != "test" {
		t.Errorf("Plugin name not set correctly: %s", candidates[1].PluginName)
	}
}

func TestPreRunCheckWithPlugins(t *testing.T) {
	registry := NewRegistry()
	
	plugin := &mockPlugin{
		name: "test",
		preRunCheckFunc: func(ctx Context, candidate *Candidate) (*CheckResult, error) {
			return &CheckResult{
				Allowed:          true,
				RequiresApproval: true,
				ApprovalMessage:  "Test approval",
			}, nil
		},
	}
	
	registry.Register(plugin, &PluginMetadata{Name: "test", Enabled: true})
	
	oldRegistry := globalRegistry
	globalRegistry = registry
	defer func() { globalRegistry = oldRegistry }()
	
	ctx := Context{
		WorkingDir: "/test",
		User:       "testuser",
		Timestamp:  time.Now(),
	}
	
	candidate := &Candidate{
		Command:    "test command",
		PluginName: "test",
	}
	
	result, err := PreRunCheckWithPlugins(ctx, candidate)
	if err != nil {
		t.Fatalf("PreRunCheckWithPlugins() error = %v", err)
	}
	
	if !result.Allowed {
		t.Error("PreRunCheckWithPlugins() should allow")
	}
	
	if !result.RequiresApproval {
		t.Error("PreRunCheckWithPlugins() should require approval")
	}
}

// Mock plugin for testing
type mockPlugin struct {
	name            string
	translateFunc   func(Context, string) ([]*Candidate, error)
	preRunCheckFunc func(Context, *Candidate) (*CheckResult, error)
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Translate(ctx Context, prompt string) ([]*Candidate, error) {
	if m.translateFunc != nil {
		return m.translateFunc(ctx, prompt)
	}
	return nil, nil
}

func (m *mockPlugin) PreRunCheck(ctx Context, candidate *Candidate) (*CheckResult, error) {
	if m.preRunCheckFunc != nil {
		return m.preRunCheckFunc(ctx, candidate)
	}
	return &CheckResult{Allowed: true}, nil
}

func (m *mockPlugin) RequiresApproval(candidate *Candidate) bool {
	return false
}

func (m *mockPlugin) Scopes() []string {
	return []string{"test"}
}
