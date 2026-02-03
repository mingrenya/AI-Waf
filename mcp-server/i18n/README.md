# MCP Server Internationalization (I18n)

This directory contains internationalization support for the AI-WAF MCP Server.

## Quick Start

### Environment Variables

```bash
# Set language (default: zh)
export WAF_LOCALE=zh   # Chinese
export WAF_LOCALE=en   # English

# Optional: Custom translation file path
export I18N_PATH=/path/to/i18n
```

### Usage Example

```go
import "github.com/mingrenya/AI-Waf/mcp-server/i18n"

// Get translation instance
i18n := i18n.GetInstance()

// Get tool translation
trans := i18n.T("list_attack_logs")
fmt.Println(trans.Name)        // Tool name
fmt.Println(trans.Description) // Tool description

// Get field translation
fieldDesc := i18n.TField("list_attack_logs", "page")
fmt.Println(fieldDesc) // Field description
```

## Translation Files

### Structure

```json
{
  "tool_name": {
    "name": "Tool Display Name",
    "description": "Detailed description of what the tool does",
    "fields": {
      "field1": "Description of field1",
      "field2": "Description of field2"
    }
  }
}
```

### Supported Languages

- **en.json**: English translations
- **zh.json**: Chinese translations (简体中文)

## Adding New Translations

### 1. Add a new language file

Create `i18n/<locale>.json`:

```json
{
  "list_attack_logs": {
    "name": "Your Translation",
    "description": "Detailed description in your language",
    "fields": {
      "page": "Page description",
      "srcIp": "Source IP description"
    }
  }
}
```

### 2. Register the locale in i18n.go

```go
const (
    LocaleEN Locale = "en"
    LocaleZH Locale = "zh"
    LocaleYourLang Locale = "xx"  // Add your language
)

func (i *I18n) loadTranslations() {
    // ... existing code ...
    i.translations[LocaleYourLang] = loadTranslationFile(translationPath + "/xx.json")
}
```

## Best Practices

### 1. Keep jsonschema tags in English

✅ **Correct**:
```go
type Input struct {
    Page int `json:"page" jsonschema:"description=Page number,default=1"`
}
```

❌ **Wrong**:
```go
type Input struct {
    Page int `json:"page" jsonschema:"页码,默认1"`  // Don't use non-ASCII
}
```

### 2. Use standard jsonschema keywords

- `description=Text`: Field description
- `enum=val1|val2`: Enumeration values
- `required`: Required field
- `default=value`: Default value
- `minimum=n,maximum=n`: Number range

### 3. Provide complete translations

- Tool name
- Tool description
- All field descriptions

## Fallback Behavior

If a translation is missing:

1. Falls back to English translation
2. If English is also missing, uses the key name
3. Ensures the server never crashes due to missing translations

## Contributing

When adding new tools:

1. Define structs with English jsonschema tags
2. Add translations to all language files
3. Test with different locales

---

For detailed documentation, see [I18N_GUIDE.md](../I18N_GUIDE.md)
