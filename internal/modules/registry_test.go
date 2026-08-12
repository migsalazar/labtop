package modules

import "testing"

func TestRegistryContainsOnlyBuiltInTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		moduleType     Type
		settings       SettingsKind
		refreshAllowed bool
	}{
		{moduleType: TypeSystem, settings: SettingsSystem, refreshAllowed: true},
		{moduleType: TypeMachines, settings: SettingsMachines, refreshAllowed: true},
		{moduleType: TypeAgents, settings: SettingsAgents},
		{moduleType: TypeEvents, settings: SettingsNone},
	}

	for _, test := range tests {
		t.Run(string(test.moduleType), func(t *testing.T) {
			t.Parallel()

			definition, ok := Lookup(test.moduleType)
			if !ok {
				t.Fatalf("Lookup(%q) returned false", test.moduleType)
			}
			if definition.Type != test.moduleType || definition.Settings != test.settings || !definition.Singleton || definition.RefreshAllowed != test.refreshAllowed {
				t.Fatalf("Lookup(%q) = %#v", test.moduleType, definition)
			}
		})
	}

	if _, ok := Lookup("unknown"); ok {
		t.Fatal("Lookup(unknown) returned true")
	}
}
