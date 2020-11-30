package cmd

// StringParameter is a parameter of type string
type StringParameter ParameterValue

func (param StringParameter) registerFlag() {
	param.FlagValue = param.Default.(string)
	if param.FlagName != "" {
		flagValue := param.FlagValue.(string)
		generateCmd.PersistentFlags().StringVar(&flagValue, param.FlagName, param.Default.(string), param.Description)
	}
}

func (param StringParameter) resolve() interface{} {
	if param.LocalValue != nil {
		return (*param.LocalValue).AsString()
	}

	if param.ParentLocalValue != nil {
		return (*param.ParentLocalValue).AsString()
	}

	return param.FlagValue
}

// DO NOT SUBMIT: Move up to top package
var workflowParameter = StringParameter{
	FlagName:    "workflow",
	LocalsName:  "atlantis_workflow",
	Description: "Use different workspace for each project. Default is use default workspace",
	Default:     "",
}
