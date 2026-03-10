package js

import (
	"encoding/json"
	"fmt"
)

// AutoFill renders a script that auto-populates form fields when a select element changes.
// selectID is the HTML ID of the select element.
// mappings maps select option values to field values: map[optionValue]map[fieldID]fieldValue.
func AutoFill(selectID string, mappings map[string]map[string]string) string {
	jsonData, err := json.Marshal(mappings)
	if err != nil {
		return fmt.Sprintf("<!-- autofill error: %s -->", err.Error())
	}

	return fmt.Sprintf(`<script>
(function(){
	var sel = document.getElementById(%q);
	if (!sel) return;
	var mappings = %s;
	sel.addEventListener("change", function() {
		var val = sel.value;
		var fields = mappings[val];
		if (!fields) return;
		for (var fieldId in fields) {
			if (!fields.hasOwnProperty(fieldId)) continue;
			var el = document.getElementById(fieldId);
			if (el) {
				el.value = fields[fieldId];
				el.dispatchEvent(new Event("input", { bubbles: true }));
			}
		}
	});
})();
</script>`, selectID, string(jsonData))
}
