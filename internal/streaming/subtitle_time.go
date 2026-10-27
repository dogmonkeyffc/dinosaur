package streaming

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var cueTime = regexp.MustCompile(`(?m)^(-?[0-9:.]+) --> (-?[0-9:.]+)([^\n]*)$`)

func normalizeSubtitle(data []byte) ([]byte, error) {
	valid := true
	result := cueTime.ReplaceAllStringFunc(string(data), func(line string) string {
		parts := cueTime.FindStringSubmatch(line)
		start, ok := subtitleSeconds(parts[1])
		end, other := subtitleSeconds(parts[2])
		if !ok || !other {
			valid = false
			return line
		}
		return subtitleClock(max(0, start)) + " --> " + subtitleClock(max(0, end)) + parts[3]
	})
	if !valid {
		return nil, ErrSubtitle
	}
	return []byte(result), nil
}
func subtitleSeconds(value string) (float64, bool) {
	sign := 1.0
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = value[1:]
	}
	parts := strings.Split(value, ":")
	if len(parts) < 2 || len(parts) > 3 {
		return 0, false
	}
	result := 0.0
	for _, part := range parts {
		n, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return 0, false
		}
		result = result*60 + n
	}
	return sign * result, true
}
func subtitleClock(seconds float64) string {
	ms := int64(seconds*1000 + 0.5)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", ms/3600000, ms/60000%60, ms/1000%60, ms%1000)
}
