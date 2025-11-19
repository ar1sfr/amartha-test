package ingest

import (
	"strconv"
	"strings"
	"time"
)

func parseCents(s string) (int64, error) {
	t := strings.TrimSpace(s)
	if t == "" {
		return 0, nil
	}
	sign := int64(1)
	if strings.HasPrefix(t, "-") {
		sign = -1
		t = t[1:]
	}
	parts := strings.SplitN(t, ".", 2)
	w := parts[0]
	f := "00"
	if len(parts) == 2 {
		f = parts[1]
	}
	if len(f) == 1 {
		f = f + "0"
	}
	if len(f) == 0 {
		f = "00"
	}
	if len(f) > 2 {
		f = f[:2]
	}
	wi, err := strconv.ParseInt(w, 10, 64)
	if err != nil {
		return 0, err
	}
	fi, err := strconv.ParseInt(f, 10, 64)
	if err != nil {
		return 0, err
	}
	return sign * (wi*100 + fi), nil
}

func parseAnyTime(s string, loc *time.Location) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15.04",
		"2006-01-02",
	}
	var lastErr error
	for _, l := range layouts {
		tm, err := time.ParseInLocation(l, s, loc)
		if err == nil {
			return tm, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}
