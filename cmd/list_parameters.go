package cmd

type ListParameter ParameterValue

func (param ListParameter) registerFlag() {
	param.FlagValue = param.Default.([]string)
	if param.FlagName != "" {
		flagValue := param.FlagValue.([]string)
		generateCmd.PersistentFlags().StringSliceVar(&flagValue, param.FlagName, param.Default.([]string), param.Description)
	}
}

// Merges everything
func (param ListParameter) resolve() interface{} {
	allValues := []string{}

	if param.LocalValue != nil {
		it := (*param.LocalValue).ElementIterator()
		for it.Next() {
			_, val := it.Element()
			allValues = append(
				allValues,
				val.AsString(),
			)
		}
	}

	if param.ParentLocalValue != nil {
		it := (*param.ParentLocalValue).ElementIterator()
		for it.Next() {
			_, val := it.Element()
			allValues = append(
				allValues,
				val.AsString(),
			)
		}
	}

	// DO NOT SUBMIT
	// for _, val := range param.FlagValue.([]string) {
	// 	allValues = append(allValues, val)
	// }

	return allValues
}

// DO NOT SUBMIT: Move up to top package
var extraDependenciesParameter = ListParameter{
	LocalsName:  "extra_atlantis_dependencies",
	Description: "Extra Atlantis Dependencies that should be added to some Project",
	Default:     []string{},
}
