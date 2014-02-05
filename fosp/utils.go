package fosp

func indexOf(array []string, element string) int {
	for i, v := range array {
		if v == element {
			return i
		}
	}
	return -1
}

func contains(array []string, element string) bool {
	return indexOf(array, element) != -1
}
