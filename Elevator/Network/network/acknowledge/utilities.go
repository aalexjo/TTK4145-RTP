package acknowledge

//Utilities for arrays in golang:
//Looks for a string in an array and returns the index.
func stringInSlice(a string, list []string) int {
	for ind, b := range list {
		if b == a {
			return ind
		}
	}
	return -1
}

//Removes a string from an array, does not care about sorting.
func removeFromSlice(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
