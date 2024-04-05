// internal/rules/rule.go

package rules

type Rule struct {
	Name          string     `json:"name"`
	Priority      int        `json:"priority"`
	Conditions    Conditions `json:"conditions"`
	Event         Event      `json:"event"`
	ProducedFacts []string   `json:"producedFacts,omitempty"` // Facts produced by this rule
	ConsumedFacts []string   `json:"consumedFacts,omitempty"` // Facts consumed by this rule
}

type Event struct {
	EventType      string        `json:"eventType"`
	CustomProperty interface{}   `json:"customProperty,omitempty"`
	Facts          []string      `json:"facts,omitempty"`
	Values         []interface{} `json:"values,omitempty"`
	Actions        []Action      `json:"actions,omitempty"`
}

type Action struct {
	Type   string      `json:"type"`   // "updateStore" or "sendMessage"
	Target string      `json:"target"` // Key for store update or address for message
	Value  interface{} `json:"value"`  // Value for store update or message content
}

type Conditions struct {
	All []Condition `json:"all,omitempty"`
	Any []Condition `json:"any,omitempty"` // `omitempty` will omit this if nil or empty
}

// Condition represents a condition used in a rule.
type Condition struct {
	Fact      string      `json:"fact"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	ValueType string      `json:"valueType,omitempty"`
	All       []Condition `json:"all,omitempty"`
	Any       []Condition `json:"any,omitempty"`
}

// RuleEngineContext holds global or shared data useful across the rules engine.
type RuleEngineContext struct {
	FactIndex     map[string]int
	ConsumedFacts map[string]bool // Tracks which facts are consumed by rules
	ProducedFacts map[string]bool // Tracks which facts are produced by rules
}

// NewRuleEngineContext initializes and returns a new RuleEngineContext.
func NewRuleEngineContext() *RuleEngineContext {
	return &RuleEngineContext{
		FactIndex:     make(map[string]int),
		ConsumedFacts: make(map[string]bool),
		ProducedFacts: make(map[string]bool),
	}
}
