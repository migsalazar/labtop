// Package modules defines the finite registry of built-in Labtop module types.
package modules

// Type identifies a built-in module implementation.
type Type string

const (
	TypeSystem   Type = "system"
	TypeMachines Type = "machines"
	TypeAgents   Type = "agents"
	TypeEvents   Type = "events"
)

// SettingsKind identifies the type-specific settings block a module requires.
type SettingsKind string

const (
	SettingsNone     SettingsKind = ""
	SettingsSystem   SettingsKind = "system"
	SettingsMachines SettingsKind = "machines"
	SettingsAgents   SettingsKind = "agents"
)

// Definition describes validation-relevant behavior for a built-in type.
type Definition struct {
	Type           Type
	Settings       SettingsKind
	Singleton      bool
	RefreshAllowed bool
}

var registry = map[Type]Definition{
	TypeSystem: {
		Type:           TypeSystem,
		Settings:       SettingsSystem,
		Singleton:      true,
		RefreshAllowed: true,
	},
	TypeMachines: {
		Type:           TypeMachines,
		Settings:       SettingsMachines,
		Singleton:      true,
		RefreshAllowed: true,
	},
	TypeAgents: {
		Type:      TypeAgents,
		Settings:  SettingsAgents,
		Singleton: true,
	},
	TypeEvents: {
		Type:      TypeEvents,
		Settings:  SettingsNone,
		Singleton: true,
	},
}

// Lookup returns the definition for a built-in module type.
func Lookup(moduleType Type) (Definition, bool) {
	definition, ok := registry[moduleType]
	return definition, ok
}
