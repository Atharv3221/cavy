package utils

func SetMaxCapacity(defaultValue int,
	newValue int, size int) (int, bool) {
	value := defaultValue
	result := false
	if newValue > 0 && newValue >= size {
		value = newValue
		result = true
	}
	return value, result
}
