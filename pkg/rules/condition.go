// pkg/rules/condition.go

package rules

const (
	OperatorEqual              = "equal"
	OperatorNotEqual           = "notEqual"
	OperatorGreaterThan        = "greaterThan"
	OperatorGreaterThanOrEqual = "greaterThanOrEqual"
	OperatorLessThan           = "lessThan"
	OperatorLessThanOrEqual    = "lessThanOrEqual"
	OperatorContains           = "contains"
	OperatorNotContains        = "notContains"
)

var SupportedOperators = []string{
	OperatorEqual,
	OperatorNotEqual,
	OperatorGreaterThan,
	OperatorGreaterThanOrEqual,
	OperatorLessThan,
	OperatorLessThanOrEqual,
	OperatorContains,
	OperatorNotContains,
}
