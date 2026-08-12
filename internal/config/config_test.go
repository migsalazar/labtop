package config

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/migsalazar/labtop/internal/modules"
)

func TestLoadMissingImplicitPathUsesIndependentDefaults(t *testing.T) {
	t.Chdir(t.TempDir())

	first, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if first.Console.Title != defaultTitle || first.Layout.Columns != 3 || len(first.Modules) != 4 {
		t.Fatalf("defaults = %#v", first)
	}
	if len(first.Network.ExternalTargets) != 0 {
		t.Fatalf("default external targets = %#v, want none", first.Network.ExternalTargets)
	}
	if len(first.Modules[2].Machines.Items) != 0 {
		t.Fatalf("default machines = %#v, want none", first.Modules[2].Machines.Items)
	}

	first.Modules[0].Title = "MUTATED"
	first.Attention.EventTypes[0] = "MUTATED"
	second, err := Load("")
	if err != nil {
		t.Fatal(err)
	}
	if second.Modules[0].Title == "MUTATED" || second.Attention.EventTypes[0] == "MUTATED" {
		t.Fatal("default configurations share mutable state")
	}
}

func TestExampleConfigurationValidates(t *testing.T) {
	path := filepath.Join("..", "..", "config.example.toml")
	configuration, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(configuration.Modules) != 4 {
		t.Fatalf("example modules = %d, want 4", len(configuration.Modules))
	}
}

func TestLoadExplicitConfiguration(t *testing.T) {
	path := writeConfig(t, minimalConfig(`title = "CUSTOM"`))
	configuration, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Console.Title != "CUSTOM" {
		t.Fatalf("title = %q, want CUSTOM", configuration.Console.Title)
	}
	if len(configuration.Modules) != 1 || configuration.Modules[0].Type != modules.TypeEvents {
		t.Fatalf("modules = %#v, want only configured events module", configuration.Modules)
	}
}

func TestLoadMissingExplicitPathFails(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.toml"))
	if err == nil || !strings.Contains(err.Error(), "read") {
		t.Fatalf("error = %v, want read error", err)
	}
}

func TestStrictDecodeRejectsMalformedAndUnknownFields(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"malformed":      "[console\ntitle = true",
		"unknown top":    "unknown = true",
		"unknown nested": "[console]\ntitel = \"TYPO\"",
	}
	for name, content := range tests {
		content := content
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := Load(writeConfig(t, content))
			if err == nil || !strings.Contains(err.Error(), "decode TOML") {
				t.Fatalf("error = %v, want decode error", err)
			}
		})
	}
}

func TestModuleValidationFailures(t *testing.T) {
	t.Parallel()

	base := minimalConfig("")
	tests := map[string]string{
		"empty modules": "modules = []\n",
		"unknown type":  strings.Replace(base, `type = "events"`, `type = "unknown"`, 1),
		"duplicate id": base + `
[[modules]]
id = "recent-events"
type = "agents"
title = "AGENTS"
column = 1
row = 0
column_span = 1
row_span = 1
[modules.agents]
provider = ""
`,
		"duplicate singleton": base + `
[[modules]]
id = "more-events"
type = "events"
title = "EVENTS"
column = 1
row = 0
column_span = 1
row_span = 1
`,
		"events settings":           base + "\n[modules.agents]\nprovider = \"\"\n",
		"events refresh":            strings.Replace(base, "row_span = 1", "row_span = 1\nrefresh_seconds = 1", 1),
		"missing required settings": strings.Replace(systemConfig(), "\n[modules.system]\n", "\n# [modules.system]\n", 1),
		"unsupported provider":      agentsConfig("provider = \"future\""),
	}
	for name, content := range tests {
		content := content
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Load(writeConfig(t, content)); err == nil {
				t.Fatal("Load() returned no error")
			}
		})
	}
}

func TestLayoutValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"negative position": strings.Replace(minimalConfig(""), "column = 0", "column = -1", 1),
		"zero span":         strings.Replace(minimalConfig(""), "column_span = 1", "column_span = 0", 1),
		"out of bounds":     strings.Replace(minimalConfig(""), "column_span = 1", "column_span = 4", 1),
		"wrong columns":     strings.Replace(minimalConfig(""), "columns = 3", "columns = 2", 1),
		"collision": minimalConfig("") + `
[[modules]]
id = "agents"
type = "agents"
title = "AGENTS"
column = 0
row = 0
column_span = 2
row_span = 1
[modules.agents]
provider = ""
`,
	}
	for name, content := range tests {
		content := content
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Load(writeConfig(t, content)); err == nil {
				t.Fatal("Load() returned no error")
			}
		})
	}
}

func TestAdjacentModulesAreSortedAndDoNotOverlap(t *testing.T) {
	configuration, err := Load(writeConfig(t, `[layout]
columns = 3

[[modules]]
id = "events"
type = "events"
title = "EVENTS"
column = 2
row = 0
column_span = 1
row_span = 1

[[modules]]
id = "agents"
type = "agents"
title = "AGENTS"
column = 0
row = 0
column_span = 2
row_span = 1
[modules.agents]
provider = ""
`))
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Modules[0].ID != "agents" || configuration.Modules[1].ID != "events" {
		t.Fatalf("module order = %q, %q", configuration.Modules[0].ID, configuration.Modules[1].ID)
	}
}

func TestNumericValidation(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"zero refresh":        strings.Replace(systemConfig(), "refresh_seconds = 1", "refresh_seconds = 0", 1),
		"negative timeout":    strings.Replace(machinesConfig(""), "probe_timeout_seconds = 1.5", "probe_timeout_seconds = -1", 1),
		"zero disk threshold": strings.Replace(systemConfig(), "disk_warning_percent = 90", "disk_warning_percent = 0", 1),
		"nonintegral history": strings.Replace(systemConfig(), "sparkline_sample_seconds = 5", "sparkline_sample_seconds = 7", 1),
		"duration overflow":   strings.Replace(minimalConfig(""), "return_home_seconds = 15", "return_home_seconds = 1e30", 1),
	}
	for name, content := range tests {
		content := content
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Load(writeConfig(t, content)); err == nil {
				t.Fatal("Load() returned no error")
			}
		})
	}

	for name, number := range map[string]float64{"NaN": math.NaN(), "positive infinity": math.Inf(1), "negative infinity": math.Inf(-1)} {
		t.Run(name, func(t *testing.T) {
			raw := defaultRawConfig()
			raw.Attention.ReturnHomeSeconds = floatPointer(number)
			if _, err := validate(raw); err == nil {
				t.Fatal("validate() returned no error")
			}
		})
	}
}

func TestMachineValidation(t *testing.T) {
	t.Parallel()

	validItem := `
[[modules.machines.items]]
id = "node-a"
label = "NODE A"
host = "192.0.2.10"
port = 12345
`
	if _, err := Load(writeConfig(t, machinesConfig(""))); err != nil {
		t.Fatalf("empty machine list: %v", err)
	}
	if _, err := Load(writeConfig(t, machinesConfig(validItem))); err != nil {
		t.Fatalf("valid machine: %v", err)
	}

	tests := map[string]string{
		"missing port":   strings.Replace(validItem, "port = 12345", "", 1),
		"zero port":      strings.Replace(validItem, "port = 12345", "port = 0", 1),
		"large port":     strings.Replace(validItem, "port = 12345", "port = 65536", 1),
		"missing host":   strings.Replace(validItem, `host = "192.0.2.10"`, `host = ""`, 1),
		"duplicate id":   validItem + strings.Replace(validItem, `label = "NODE A"`, `label = "NODE B"`, 1),
		"three machines": validItem + strings.Replace(validItem, `id = "node-a"`, `id = "node-b"`, 1) + strings.Replace(validItem, `id = "node-a"`, `id = "node-c"`, 1),
	}
	for name, items := range tests {
		items := items
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Load(writeConfig(t, machinesConfig(items))); err == nil {
				t.Fatal("Load() returned no error")
			}
		})
	}
}

func TestExternalTargetsAndAttentionAllowlist(t *testing.T) {
	t.Parallel()

	validTarget := minimalConfig("") + `
[[network.external_targets]]
label = "EXAMPLE"
host = "203.0.113.10"
port = 443
`
	configuration, err := Load(writeConfig(t, validTarget))
	if err != nil {
		t.Fatal(err)
	}
	if len(configuration.Network.ExternalTargets) != 1 || configuration.Network.ExternalTargets[0].Port != 443 {
		t.Fatalf("targets = %#v", configuration.Network.ExternalTargets)
	}

	for name, content := range map[string]string{
		"invalid target":  strings.Replace(validTarget, "port = 443", "port = 0", 1),
		"unknown event":   strings.Replace(minimalConfig(""), `"machine.offline"`, `"unknown.event"`, 1),
		"duplicate event": strings.Replace(minimalConfig(""), `["machine.offline"]`, `["machine.offline", "machine.offline"]`, 1),
	} {
		content := content
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Load(writeConfig(t, content)); err == nil {
				t.Fatal("Load() returned no error")
			}
		})
	}
}

func TestOmittedValuesReceiveDefaults(t *testing.T) {
	configuration, err := Load(writeConfig(t, `[layout]
columns = 3
[[modules]]
id = "system"
type = "system"
title = "SYSTEM"
column = 0
row = 0
column_span = 1
row_span = 1
[modules.system]
`))
	if err != nil {
		t.Fatal(err)
	}
	module := configuration.Modules[0]
	if module.Refresh != time.Second || module.System.SparklineCapacity != 180 || module.System.SlowRefresh != 5*time.Second {
		t.Fatalf("defaulted system module = %#v", module)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func minimalConfig(consoleFields string) string {
	return `[console]
` + consoleFields + `
[layout]
columns = 3
[attention]
return_home_seconds = 15
event_types = ["machine.offline"]
[network]
external_probe_seconds = 15
external_probe_timeout_seconds = 2
[[modules]]
id = "recent-events"
type = "events"
title = "EVENTS"
column = 0
row = 0
column_span = 1
row_span = 1
`
}

func systemConfig() string {
	return `[layout]
columns = 3
[[modules]]
id = "system"
type = "system"
title = "SYSTEM"
column = 0
row = 0
column_span = 1
row_span = 1
refresh_seconds = 1
[modules.system]
local_label = "LOCAL NODE"
slow_refresh_seconds = 5
sparkline_minutes = 15
sparkline_sample_seconds = 5
temperature_warning_celsius = 80
disk_warning_percent = 90
`
}

func agentsConfig(settings string) string {
	return `[layout]
columns = 3
[[modules]]
id = "agents"
type = "agents"
title = "AGENTS"
column = 0
row = 0
column_span = 1
row_span = 1
[modules.agents]
` + settings + "\n"
}

func machinesConfig(items string) string {
	return `[layout]
columns = 3
[[modules]]
id = "machines"
type = "machines"
title = "MACHINES"
column = 0
row = 0
column_span = 1
row_span = 1
refresh_seconds = 5
[modules.machines]
probe_timeout_seconds = 1.5
` + items
}
