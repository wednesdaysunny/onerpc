package mobile

import (
	"strconv"
	"strings"
)

// VersionCompare compares two versions,
// returns 1 when ver1 is greater than ver2, -1 when ver1 is less than ver2, otherwise 0.
func VersionCompare(ver1 string, ver2 string, sep string) int {
	parts1 := strings.Split(ver1, sep)
	parts2 := strings.Split(ver2, sep)
	minLen := len(parts1)
	if len(parts2) < minLen {
		minLen = len(parts2)
	}

	for i := 0; i < minLen; i++ {
		num1, err := strconv.ParseInt(parts1[i], 10, 64)
		if err != nil {
			return 1
		}
		num2, err := strconv.ParseInt(parts2[i], 10, 64)
		if err != nil {
			return 1
		}
		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}

	if len(parts1) > len(parts2) {
		return 1
	} else if len(parts1) < len(parts2) {
		return -1
	}
	return 0
}
