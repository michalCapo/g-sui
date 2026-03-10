package js

import "fmt"

// AsyncButton renders a button that POSTs to a URL on click, shows loading state,
// and displays the result in a target element.
func AsyncButton(label, url, resultID, class string) string {
	if class == "" {
		class = "bg-blue-600 text-white px-4 py-2 rounded text-sm hover:bg-blue-700 cursor-pointer"
	}
	btnID := "abtn_" + resultID
	return fmt.Sprintf(`<button type="button" id="%s" class="%s">%s</button>
<script>
(function(){
	var btn = document.getElementById(%q);
	if (!btn) return;
	var origText = btn.textContent;
	btn.addEventListener("click", function() {
		btn.disabled = true;
		btn.textContent = "Loading...";
		btn.style.opacity = "0.6";

		var form = btn.closest("form");
		var body = {};
		if (form) {
			var fd = new FormData(form);
			fd.forEach(function(v, k) { body[k] = v; });
		}

		fetch(%q, {
			method: "POST",
			headers: { "Content-Type": "application/json", "Accept": "application/json" },
			credentials: "same-origin",
			body: JSON.stringify(body)
		})
		.then(function(resp) {
			if (!resp.ok) throw new Error("HTTP " + resp.status);
			return resp.json();
		})
		.then(function(data) {
			btn.disabled = false;
			btn.textContent = origText;
			btn.style.opacity = "1";
			var res = document.getElementById(%q);
			if (res) {
				res.textContent = data.message || JSON.stringify(data);
				res.className = "text-sm text-green-600 dark:text-green-400 mt-2";
			}
		})
		.catch(function(err) {
			btn.disabled = false;
			btn.textContent = origText;
			btn.style.opacity = "1";
			var res = document.getElementById(%q);
			if (res) {
				res.textContent = err.message || "Request failed";
				res.className = "text-sm text-red-600 dark:text-red-400 mt-2";
			}
		});
	});
})();
</script>
<div id="%s"></div>`, btnID, class, label, btnID, url, resultID, resultID, resultID)
}
