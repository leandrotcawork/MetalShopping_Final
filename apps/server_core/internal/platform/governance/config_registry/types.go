package config_registry

type Scope string

const (
	ScopeGlobal        Scope = "global"
	ScopeEnvironment   Scope = "environment"
	ScopeTenant        Scope = "tenant"
	ScopeModule        Scope = "module"
	ScopeEntityProfile Scope = "entity/profile"
	ScopeFeatureTarget Scope = "feature-target"
)

type ArtifactKind string

const (
	ArtifactFeatureFlag ArtifactKind = "feature_flag"
	ArtifactThreshold   ArtifactKind = "threshold"
	ArtifactPolicy      ArtifactKind = "policy"
)

type ValueType string

const (
	ValueTypeBool   ValueType = "bool"
	ValueTypeNumber ValueType = "number"
	ValueTypeJSON   ValueType = "json"
)

type Entry struct {
	Key            string
	Kind           ArtifactKind
	BoundedContext string
	ValueType      ValueType
	Scopes         []Scope
	Description    string
}
