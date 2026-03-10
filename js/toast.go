package js

import "fmt"

// Toast renders a script that shows a toast notification via __notify.
// This is a convenience for server-side triggered toasts.
func Toast(message, variant string) string {
	if variant == "" {
		variant = "success"
	}
	return fmt.Sprintf(`<script>
if (typeof __notify === "function") __notify(%q, %q);
else if (typeof __toast !== "undefined") __toast.show(%q, %q);
</script>`, message, variant, message, variant)
}
