package linear

// removeItemByIndex remove item out of []string by index but maintains order, and return the new one
// Source: https://yourbasic.org/golang/delete-element-slice/
func removeItemByIndex(slice []string, idx int) []string {

	copy(slice[idx:], slice[idx+1:]) // Shift slice[idx+1:] left one index.
	slice[len(slice)-1] = ""         // Erase last element (write zero value).
	return slice[:len(slice)-1]      // Truncate slice.
}

// findIndexByItem return index belong to the key
// Source: https://stackoverflow.com/questions/46745043/performance-of-for-range-in-go
func findIndexByItem(keyName string, items []string) (int, bool) {

	for index := range items {
		if keyName == items[index] {
			return index, true
		}
	}

	return -1, false
}
