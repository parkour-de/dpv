package t

import (
	"dpv/dpv/src/repository/dpv"
	"fmt"
	"os"
	"strings"
)

// TranslatableError captures the intent to translate.
type TranslatableError struct {
	Key  string
	Args []any
}

// Error implements the standard error interface with a fallback (e.g. English).
func (e *TranslatableError) Error() string {
	// Fallback logic: raw key + formatted args
	return fmt.Sprintf(e.Key, e.Args...)
}

// Unwrap allows standard errors.Is/As checks to work by unwrapping the *first* error found in Args.
func (e *TranslatableError) Unwrap() error {
	for _, arg := range e.Args {
		if err, ok := arg.(error); ok {
			return err
		}
	}
	return nil
}

// Errorf creates the structured error.
func Errorf(key string, args ...any) error {
	return &TranslatableError{
		Key:  key,
		Args: args,
	}
}

// Translate recursively translates a TranslatableError and its nested errors.
func Translate(err error, langMap map[string]string) string {
	if err == nil {
		return ""
	}

	if tErr, ok := err.(*TranslatableError); ok {
		translatedArgs := make([]any, len(tErr.Args))
		for i, arg := range tErr.Args {
			if argErr, isErr := arg.(error); isErr {
				translatedArgs[i] = Translate(argErr, langMap)
			} else {
				translatedArgs[i] = arg
			}
		}

		format, exists := langMap[tErr.Key]
		if !exists {
			format = tErr.Key
		}

		// Replace %w with %s because we are passing translated strings (recursive step)
		format = strings.ReplaceAll(format, "%w", "%s")

		if len(translatedArgs) == 0 {
			return format
		}
		return fmt.Sprintf(format, translatedArgs...)
	}

	return err.Error()
}

var languages = make(map[string]map[string]string)

func GetMapFor(lang string) map[string]string {
	if m, ok := languages[lang]; ok {
		return m
	}
	// Fallback to parts of the language tag (e.g. de-DE -> de)
	if idx := strings.Index(lang, "-"); idx != -1 {
		if m, ok := languages[lang[:idx]]; ok {
			return m
		}
	}
	return make(map[string]string)
}

func T(err error, lang string) string {
	return Translate(err, GetMapFor(lang))
}

func LoadLanguages(config *dpv.Config) error {
	if config == nil {
		return fmt.Errorf("config is not initialized")
	}

	for _, lang := range config.Settings.SupportedLanguages {
		// English uses the raw keys as readable fallbacks and does not require a
		// translation file. Skipping prevents noisy warnings about a missing
		// strings_en.ini while still allowing "en" as a valid language choice.
		if lang == "en" {
			continue
		}

		path := config.Path + "strings_" + lang + ".ini"
		m, err := loadFile(path)
		if err != nil {
			fmt.Printf("Warning: could not load %s: %v\n", path, err)
			continue
		}
		languages[lang] = m
	}
	return nil
}

func loadFile(path string) (map[string]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, s := range strings.Split(string(bytes), "\n") {
		s = strings.TrimSpace(s)
		if s == "" || strings.HasPrefix(s, ";") {
			continue
		}
		arr := strings.SplitN(s, "=", 2)
		if len(arr) != 2 {
			continue
		}
		key := strings.TrimSpace(arr[0])
		val := strings.TrimSpace(arr[1])
		if val != "" {
			m[key] = val
		}
	}
	return m, nil
}
