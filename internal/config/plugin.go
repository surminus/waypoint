package config

import (
	"github.com/hashicorp/waypoint/sdk/component"
)

// Plugins returns all the plugins defined by this configuration. This
// will include the implicitly defined plugins via `use` statements.
func (c *Config) Plugins() []*Plugin {
	var result []*Plugin
	known := map[string]*Plugin{}

	// Collect all the plugins used by all the apps.
	for _, app := range c.Apps {
		// Get all the implied stage plugins: build, deploy, etc.
		if v := app.Build; v != nil {
			result = trackPlugin(result, known, v.Use, component.BuilderType)
			if v := v.Registry; v != nil {
				result = trackPlugin(result, known, v.Use, component.RegistryType)
			}
		}
		if v := app.Deploy; v != nil {
			result = trackPlugin(result, known, v.Use, component.PlatformType)
		}
		if v := app.Release; v != nil {
			result = trackPlugin(result, known, v.Use, component.ReleaseManagerType)
		}
	}

	return result
}

// trackPlugin adds the plugin implied by the use statements to result if
// it hasn't been seen before (known via the "known" variable). This will
// modify the values result and known in-place.
func trackPlugin(
	result []*Plugin,
	known map[string]*Plugin,
	use *Use,
	typ component.Type,
) []*Plugin {
	p, ok := known[use.Type]
	if !ok {
		p = &Plugin{Name: use.Type}
		result = append(result, p)
		known[use.Type] = p
	}

	p.markType(typ)
	return result
}

// Plugin configures a plugin.
type Plugin struct {
	// Name of the plugin. This is expected to match the plugin binary
	// "waypoint-plugin-<name>" including casing.
	Name string `hcl:",label"`

	// Type is the type of plugin this is. This can be multiple.
	Type struct {
		Builder  bool `hcl:"build,optional"`
		Registry bool `hcl:"registry,optional"`
		Platform bool `hcl:"deploy,optional"`
		Releaser bool `hcl:"release,optional"`
	} `hcl:"type,block"`

	// Checksum is the SHA256 checksum to validate this plugin.
	Checksum string `hcl:"checksum,optional"`
}

// Types returns the list of types that this plugin implements.
func (p *Plugin) Types() []component.Type {
	var result []component.Type
	for t, b := range p.typeMap() {
		if *b {
			result = append(result, t)
		}
	}

	return result
}

// markType marks that the given component type is supported by this plugin.
// This will panic if an unsupported plugin type is given.
func (p *Plugin) markType(typ component.Type) {
	m := p.typeMap()
	b, ok := m[typ]
	if !ok {
		panic("unknown type: " + typ.String())
	}

	*b = true
}

func (p *Plugin) typeMap() map[component.Type]*bool {
	return map[component.Type]*bool{
		component.BuilderType:        &p.Type.Builder,
		component.RegistryType:       &p.Type.Registry,
		component.PlatformType:       &p.Type.Platform,
		component.ReleaseManagerType: &p.Type.Releaser,
	}
}
