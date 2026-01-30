package asc

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// RateLimitInfo captures parsed rate limit header information.
//
// Apple has exposed an X-Rate-Limit header in formats like:
//   user-hour-lim:<n>;user-hour-rem:<m>;
// (see Apple Dev Forums thread 110457).
type RateLimitInfo struct {
	Windows map[string]RateLimitWindow
	Raw     string
}

type RateLimitWindow struct {
	Limit     *int
	Remaining *int
}

func (r *RateLimitInfo) Summary() string {
	if r == nil || len(r.Windows) == 0 {
		return ""
	}
	keys := make([]string, 0, len(r.Windows))
	for k := range r.Windows {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, window := range keys {
		w := r.Windows[window]
		switch {
		case w.Limit != nil && w.Remaining != nil:
			parts = append(parts, fmt.Sprintf("%s %d/%d remaining", window, *w.Remaining, *w.Limit))
		case w.Remaining != nil:
			parts = append(parts, fmt.Sprintf("%s %d remaining", window, *w.Remaining))
		case w.Limit != nil:
			parts = append(parts, fmt.Sprintf("%s limit %d", window, *w.Limit))
		}
	}
	return strings.Join(parts, "; ")
}

func parseRateLimitHeader(value string) *RateLimitInfo {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil
	}

	info := &RateLimitInfo{
		Windows: map[string]RateLimitWindow{},
		Raw:     raw,
	}

	// Split on common separators. Header examples use ';' separators and ':' key/value.
	s := strings.NewReplacer(";", ",", "\n", ",").Replace(raw)
	for _, token := range strings.Split(s, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		sep := ":"
		if strings.Contains(token, "=") && !strings.Contains(token, ":") {
			sep = "="
		}
		key, val, ok := strings.Cut(token, sep)
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" || val == "" {
			continue
		}

		n, err := strconv.Atoi(val)
		if err != nil {
			continue
		}

		switch {
		case strings.HasSuffix(key, "-lim"):
			window := strings.TrimSuffix(key, "-lim")
			w := info.Windows[window]
			w.Limit = intPtr(n)
			info.Windows[window] = w
		case strings.HasSuffix(key, "-rem"):
			window := strings.TrimSuffix(key, "-rem")
			w := info.Windows[window]
			w.Remaining = intPtr(n)
			info.Windows[window] = w
		}
	}

	if len(info.Windows) == 0 {
		return nil
	}
	return info
}

func intPtr(v int) *int { return &v }

