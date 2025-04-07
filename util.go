package linear

func removeItemByIndex(slice []string, idx int) []string {
	if idx < 0 || idx >= len(slice) {
		// Invalid index, return original slice unchanged
		return slice
	}

	copy(slice[idx:], slice[idx+1:]) // Shift slice[idx+1:] left one index.
	slice[len(slice)-1] = ""         // Erase last element (write zero value).
	return slice[:len(slice)-1]      // Truncate slice.
}

func findIndexByItem(keyName string, items []string) (int, bool) {

	for index := range items {
		if keyName == items[index] {
			return index, true
		}
	}

	return -1, false
}
