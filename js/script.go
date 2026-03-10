package js

// Script wraps arbitrary JavaScript in a self-executing IIFE <script> block.
// The body is the raw JS code — it will be placed inside (function(){ ... })();
func Script(body string) string {
	return "<script>\n(function(){\n" + body + "\n})();\n</script>"
}
