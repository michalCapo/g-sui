package js

import "fmt"

// Autocomplete renders a datalist connected to an input element for browser-native autocomplete.
// It fetches options from sourceURL on DOMContentLoaded and populates the datalist.
func Autocomplete(inputID, sourceURL, class string) string {
	listID := inputID + "_list"
	if class == "" {
		class = "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm"
	}
	return fmt.Sprintf(`<input type="text" id="%s" list="%s" class="%s" autocomplete="off">
<datalist id="%s"></datalist>
<script>
(function(){
	var inputId = %q;
	var listId = %q;
	var url = %q;

	function loadOptions() {
		fetch(url, { method: "GET", headers: { "Accept": "application/json" }, credentials: "same-origin" })
			.then(function(r) { return r.json(); })
			.then(function(data) {
				var dl = document.getElementById(listId);
				if (!dl) return;
				dl.innerHTML = "";
				var items = Array.isArray(data) ? data : (data.items || data.options || []);
				for (var i = 0; i < items.length; i++) {
					var opt = document.createElement("option");
					if (typeof items[i] === "string") {
						opt.value = items[i];
					} else {
						opt.value = items[i].value || items[i].label || String(items[i]);
						if (items[i].label && items[i].label !== items[i].value) opt.label = items[i].label;
					}
					dl.appendChild(opt);
				}
			})
			.catch(function(e) { console.error("Autocomplete load error:", e); });
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", loadOptions);
	} else {
		loadOptions();
	}
})();
</script>`, inputID, listID, class, listID, inputID, listID, sourceURL)
}
