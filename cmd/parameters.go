package cmd

import (
	"github.com/zclconf/go-cty/cty"
)

// Parameter is a value that can be a flag, local, or both
type Parameter interface {
	// Register a flag for this parameter, if this parameter has a flag
	registerFlag()

	// Resolve to a final value from all of the parsed values
	resolve() interface{}
}

// ParameterValue is a value that will implement the Parameter inferface
type ParameterValue struct {
	// The flag for this parameter will have this name. If this param doesn't have a flag, set to empty
	FlagName string

	// The `locals` block field for this parameter will have this name. If this param can't be set in `locals`, set to empty
	LocalsName string

	// Description of the parameter
	Description string

	// The default Value
	Default interface{}

	// The result of parsing the flag
	FlagValue interface{}

	// The value from the `locals` block of the terragrunt module
	LocalValue *cty.Value

	// The value from the `locals` block of the parent terragrunt module
	ParentLocalValue *cty.Value
}

func (param ParameterValue) reset() {
	param.FlagValue = param.Default
	param.LocalValue = nil
	param.ParentLocalValue = nil
}
