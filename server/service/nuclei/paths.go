package nuclei

import (
	"os"
	"path/filepath"
)

// ResolveTemplatesPath returns the first existing nuclei templates directory.
// If no known directory exists, it returns the best configured candidate and false.
func ResolveTemplatesPath() (string, bool) {
	candidates := []string{}
	if envPath := os.Getenv("NUCLEI_TEMPLATES_PATH"); envPath != "" {
		candidates = append(candidates, envPath)
	}

	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		candidates = append(candidates, filepath.Join(homeDir, ".config", "nuclei", "nuclei-templates"))
	}

	candidates = append(candidates,
		"./nuclei-templates",
		"/app/nuclei-templates",
		"/home/mrya/mrya-waf/nuclei-templates",
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
