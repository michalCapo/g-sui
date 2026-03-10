package js

import "fmt"

// SPA renders a container div that can be replaced via client-side fetch navigation.
// id is the HTML ID for the container; content is the initial HTML inside.
func SPA(id, class string) string {
	if class == "" {
		class = ""
	}
	clsAttr := ""
	if class != "" {
		clsAttr = fmt.Sprintf(` class="%s"`, class)
	}
	return fmt.Sprintf(`<div id="%s"%s></div>
<script>
(function(){
	if (!window.__spa) {
		window.__spa = {
			history: {},
			load: function(id, url, pushState) {
				var el = document.getElementById(id);
				if (!el) return Promise.reject(new Error("SPA container not found: " + id));
				el.style.opacity = "0.5";
				return fetch(url, {
					method: "GET",
					headers: { "Accept": "text/html" },
					credentials: "same-origin"
				}).then(function(resp) {
					if (!resp.ok) throw new Error("HTTP " + resp.status);
					return resp.text();
				}).then(function(html) {
					el.innerHTML = html;
					el.style.opacity = "1";
					window.__spa.history[id] = url;
					if (pushState !== false) {
						var title = document.title;
						var titleMatch = html.match(/<title[^>]*>([^<]+)<\/title>/i);
						if (titleMatch) { title = titleMatch[1]; document.title = title; }
						history.pushState({ __spa: true, id: id, url: url }, title, url);
					}
					// Execute any inline scripts in the loaded HTML
					var scripts = el.querySelectorAll("script");
					for (var i = 0; i < scripts.length; i++) {
						var s = document.createElement("script");
						if (scripts[i].src) { s.src = scripts[i].src; }
						else { s.textContent = scripts[i].textContent; }
						scripts[i].parentNode.replaceChild(s, scripts[i]);
					}
				}).catch(function(err) {
					el.style.opacity = "1";
					if (typeof __notify === "function") __notify(err.message, "error");
					throw err;
				});
			}
		};

		window.addEventListener("popstate", function(e) {
			if (e.state && e.state.__spa) {
				window.__spa.load(e.state.id, e.state.url, false);
			}
		});
	}
})();
</script>`, id, clsAttr)
}

// SPALink renders an anchor that triggers SPA navigation on click.
// target is the SPA container ID; url is the page to load.
// Falls back to normal navigation if JS fails.
func SPALink(target, url, class string, content string) string {
	if class == "" {
		class = "text-blue-600 dark:text-blue-400 hover:underline"
	}
	return fmt.Sprintf(`<a href="%s" class="%s" onclick="event.preventDefault();if(window.__spa)window.__spa.load('%s','%s');">%s</a>`,
		url, class, target, url, content)
}
