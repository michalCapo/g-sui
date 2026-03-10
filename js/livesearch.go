package js

import "fmt"

// LiveSearch renders a search input that filters server-rendered HTML elements
// by their text content. Elements matching the query are shown; others are hidden.
// targetSelector is a CSS selector for items to filter (e.g., ".list-item").
// inputID is the HTML ID for the search input element.
func LiveSearch(targetSelector, inputID, class string) string {
	if class == "" {
		class = "bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded px-3 py-2 text-sm w-64"
	}
	return fmt.Sprintf(`<input type="text" id="%s" class="%s" placeholder="Search..." autocomplete="off">
<script>
(function(){
	var inp = document.getElementById(%q);
	if (!inp) return;
	inp.addEventListener("input", function(){
		var q = inp.value.toLowerCase();
		var items = document.querySelectorAll(%q);
		for (var i = 0; i < items.length; i++) {
			var text = items[i].textContent.toLowerCase();
			items[i].style.display = (!q || text.indexOf(q) >= 0) ? "" : "none";
		}
	});
})();
</script>`, inputID, class, inputID, targetSelector)
}
