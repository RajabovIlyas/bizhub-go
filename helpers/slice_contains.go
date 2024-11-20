package helpers

func SliceContains(s []string, element string) bool {
	for _, value := range s {
		if value == element {
			return true
		}
	}
	return false
}

func SliceContainsBy[T comparable](s []T, element T) bool {
	for _, value := range s {
		if value == element {
			return true
		}
	}
	return false
}
