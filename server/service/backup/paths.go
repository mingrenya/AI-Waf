package backup

import (
	"os"
	"path/filepath"
)

// ResolveBackupPath returns the first existing backup data directory.
// If no known directory exists, it returns the best configured candidate and false.
func ResolveBackupPath() (string, bool) {
	candidates := []string{}
	if envPath := os.Getenv("BACKUP_DATA_PATH"); envPath != "" {
		candidates = append(candidates, envPath)
	}

	// 当前工作目录下的 data/backups
	if wd, err := os.Getwd(); err == nil && wd != "" {
		candidates = append(candidates, filepath.Join(wd, "data", "backups"))
	}

	candidates = append(candidates,
		"./data/backups",
		"/app/data/backups",
		"/tmp/ai-waf/backups",
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
