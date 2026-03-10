package js

import "fmt"

// ExternalLink renders an anchor element that navigates via window.location.href,
// preventing SPA frameworks from intercepting the navigation.
func ExternalLink(url, class string, content string) string {
	if class == "" {
		class = "text-blue-600 dark:text-blue-400 hover:underline"
	}
	return fmt.Sprintf(`<a href="%s" class="%s" onclick="event.preventDefault();window.location.href='%s';">%s</a>`,
		url, class, url, content)
}
