package naming

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var re = regexp.MustCompile("^[a-z](-?[a-z0-9]+)*$")

const DelimiterDefault = "-"

type MergeMode int

const (
	ModeLower MergeMode = iota
	ModeUpper
	ModeLowerCamelCase
	ModeUpperCamelCase
)

func Validate(name string) bool {
	return re.Match([]byte(name))
}

func SplitIntoParts(name string) []string {
	return strings.Split(name, DelimiterDefault)
}

func MergeToCanonical(parts []string) string {
	return Merge(parts, DelimiterDefault, ModeLower)
}

func NameToEnv(name string) string {
	return Merge(
		append([]string{"chill", "service"}, SplitIntoParts(name)...),
		"_",
		ModeUpper,
	)
}

func SecretToEnv(key string) string {
	return Merge(
		append([]string{"chill", "secret"}, SplitIntoParts(key)...),
		"_",
		ModeUpper,
	)
}

func SecretToMountPath(key string) string {
	return fmt.Sprintf("/etc/chill/secret/%s", key)
}

func capitalizeFirst(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func Merge(parts []string, delimiter string, mode MergeMode) string {
	for i := 0; i < len(parts); i++ {
		switch mode {
		case ModeUpperCamelCase:
			parts[i] = capitalizeFirst(parts[i])
		case ModeLowerCamelCase:
			if i > 0 {
				parts[i] = capitalizeFirst(parts[i])
			}
		case ModeUpper:
			parts[i] = strings.ToTitle(parts[i])
		case ModeLower:
			break
		}
	}
	return strings.Join(parts, delimiter)
}
