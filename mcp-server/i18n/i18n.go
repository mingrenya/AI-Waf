// Package i18n provides internationalization support for MCP tools
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Locale represents a supported language
type Locale string

const (
	LocaleEN Locale = "en"
	LocaleZH Locale = "zh"
)

// Translation contains localized strings for a tool or field
type Translation struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Fields      map[string]string `json:"fields,omitempty"`
}

// I18n manages internationalization
type I18n struct {
	currentLocale Locale
	translations  map[Locale]map[string]Translation
	mu            sync.RWMutex
}

var (
	instance *I18n
	once     sync.Once
)

// GetInstance returns the singleton I18n instance
func GetInstance() *I18n {
	once.Do(func() {
		instance = &I18n{
			currentLocale: LocaleZH, // Default to Chinese
			translations:  make(map[Locale]map[string]Translation),
		}
		instance.loadTranslations()
	})
	return instance
}

// SetLocale changes the current locale
func (i *I18n) SetLocale(locale Locale) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.currentLocale = locale
}

// GetLocale returns the current locale
func (i *I18n) GetLocale() Locale {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.currentLocale
}

// GetToolTranslation returns the translation for a tool
func (i *I18n) GetToolTranslation(toolName string) Translation {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if trans, ok := i.translations[i.currentLocale][toolName]; ok {
		return trans
	}

	// Fallback to English
	if trans, ok := i.translations[LocaleEN][toolName]; ok {
		return trans
	}

	return Translation{
		Name:        toolName,
		Description: toolName,
		Fields:      make(map[string]string),
	}
}

// GetFieldTranslation returns the translation for a specific field
func (i *I18n) GetFieldTranslation(toolName, fieldName string) string {
	trans := i.GetToolTranslation(toolName)
	if desc, ok := trans.Fields[fieldName]; ok {
		return desc
	}
	return fieldName
}

// loadTranslations loads all translation files
func (i *I18n) loadTranslations() {
	// Load from environment variable or default path
	translationPath := os.Getenv("I18N_PATH")
	if translationPath == "" {
		translationPath = "./i18n"
	}

	// Load English translations
	i.translations[LocaleEN] = loadTranslationFile(translationPath + "/en.json")

	// Load Chinese translations
	i.translations[LocaleZH] = loadTranslationFile(translationPath + "/zh.json")
}

// loadTranslationFile loads a single translation file
func loadTranslationFile(path string) map[string]Translation {
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return empty map
		return make(map[string]Translation)
	}

	var translations map[string]Translation
	if err := json.Unmarshal(data, &translations); err != nil {
		fmt.Printf("Warning: failed to parse translation file %s: %v\n", path, err)
		return make(map[string]Translation)
	}

	return translations
}

// T is a convenience function for translation
func T(toolName string) Translation {
	return GetInstance().GetToolTranslation(toolName)
}

// TField is a convenience function for field translation
func TField(toolName, fieldName string) string {
	return GetInstance().GetFieldTranslation(toolName, fieldName)
}
