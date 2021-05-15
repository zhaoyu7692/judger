package utils

func Min(v1, v2 int) int {
	if v1 < v2 {
		return v1
	}
	return v2
}

func Max(v1, v2 int) int {
	if v1 < v2 {
		return v2
	}
	return v1
}
