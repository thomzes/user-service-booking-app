package error

func ErrorMapping(err error) bool {
	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...) // Fixed: added allErrors as first argument
	allErrors = append(allErrors, UserErrors...)    // Fixed: added allErrors as first argument

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
