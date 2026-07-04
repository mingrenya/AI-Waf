package capture

import (
	"os"
	"path/filepath"
)

// ResolveCapturePath returns the first existing capture data directory.
// If no known directory exists, it returns the best configured candidate and false.
func ResolveCapturePath() (string, bool) {
	candidates := []string{}
	if envPath := os.Getenv("CAPTURE_DATA_PATH"); envPath != "" {
		candidates = append(candidates, envPath)
	}

	// 当前工作目录下的 data/captures
	if wd, err := os.Getwd(); err == nil && wd != "" {
		candidates = append(candidates, filepath.Join(wd, "data", "captures"))
	}

	candidates = append(candidates,
		"./data/captures",
		"/app/data/captures",
		"/tmp/ai-waf/captures",
	)

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, true
		}
	}

	for _, candidate := range candidates {
		if candidate != "" {
			return candidate, false
		}
	}

	return "", false
}
