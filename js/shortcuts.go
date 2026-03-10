package js

import "fmt"

// Shortcuts renders a keyboard shortcut registration system.
// It supports single-key, modifier, and sequence shortcuts.
// Call this once in the layout to initialize the system.
func Shortcuts() string {
	return `<script>
(function(){
	if (window.__shortcuts) return;
	var registered = [];
	var seq = "";
	var seqTimer = null;
	var seqTimeout = 800;

	function isTyping() {
		var ae = document.activeElement;
		if (!ae) return false;
		var tag = ae.tagName;
		return tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT" || ae.isContentEditable;
	}

	function flash(el) {
		if (!el) return;
		el.style.transition = "background-color 0.15s ease";
		var orig = el.style.backgroundColor;
		el.style.backgroundColor = "rgba(59,130,246,0.2)";
		setTimeout(function() { el.style.backgroundColor = orig; }, 300);
	}

	function matchKey(e, key) {
		var parts = key.toLowerCase().split("+");
		var mainKey = parts[parts.length - 1];
		var needCtrl = parts.indexOf("ctrl") >= 0;
		var needShift = parts.indexOf("shift") >= 0;
		var needAlt = parts.indexOf("alt") >= 0;
		var needMeta = parts.indexOf("meta") >= 0 || parts.indexOf("cmd") >= 0;
		if (needCtrl !== e.ctrlKey) return false;
		if (needShift !== e.shiftKey) return false;
		if (needAlt !== e.altKey) return false;
		if (needMeta !== e.metaKey) return false;
		return e.key.toLowerCase() === mainKey || e.code.toLowerCase() === "key" + mainKey;
	}

	window.__shortcuts = {
		register: function(key, handler, description) {
			registered.push({ key: key, handler: handler, description: description || key, isSequence: key.length > 1 && key.indexOf("+") < 0 });
		},
		unregister: function(key) {
			registered = registered.filter(function(r) { return r.key !== key; });
		},
		list: function() {
			return registered.map(function(r) { return { key: r.key, description: r.description }; });
		},
		flash: flash
	};

	document.addEventListener("keydown", function(e) {
		if (isTyping()) return;

		// Check sequences first
		seq += e.key.toLowerCase();
		if (seqTimer) clearTimeout(seqTimer);
		seqTimer = setTimeout(function() { seq = ""; }, seqTimeout);

		for (var i = 0; i < registered.length; i++) {
			var r = registered[i];
			if (r.isSequence && seq.endsWith(r.key.toLowerCase())) {
				e.preventDefault();
				seq = "";
				r.handler(e);
				return;
			}
		}

		// Check single keys and modifier combos
		for (var i = 0; i < registered.length; i++) {
			var r = registered[i];
			if (!r.isSequence && matchKey(e, r.key)) {
				e.preventDefault();
				r.handler(e);
				return;
			}
		}
	});

	// Register ? for help
	window.__shortcuts.register("?", function() {
		var items = window.__shortcuts.list();
		var overlay = document.getElementById("__shortcut_help__");
		if (overlay) { overlay.remove(); return; }
		overlay = document.createElement("div");
		overlay.id = "__shortcut_help__";
		var isDark = document.documentElement.classList.contains("dark");
		overlay.style.cssText = "position:fixed;inset:0;z-index:10000;display:flex;align-items:center;justify-content:center;background:rgba(0,0,0,0.5);";
		var modal = document.createElement("div");
		modal.style.cssText = "background:" + (isDark ? "#1f2937" : "#fff") + ";border-radius:12px;padding:24px;min-width:320px;max-width:500px;max-height:70vh;overflow:auto;box-shadow:0 20px 60px rgba(0,0,0,0.2);color:" + (isDark ? "#e5e7eb" : "#111827") + ";";
		var title = document.createElement("h3");
		title.textContent = "Keyboard Shortcuts";
		title.style.cssText = "margin:0 0 16px;font-size:16px;font-weight:600;";
		modal.appendChild(title);
		for (var i = 0; i < items.length; i++) {
			var row = document.createElement("div");
			row.style.cssText = "display:flex;justify-content:space-between;padding:6px 0;border-bottom:1px solid " + (isDark ? "#374151" : "#f3f4f6") + ";";
			var desc = document.createElement("span");
			desc.style.cssText = "font-size:14px;";
			desc.textContent = items[i].description;
			var kbd = document.createElement("kbd");
			kbd.style.cssText = "background:" + (isDark ? "#374151" : "#f3f4f6") + ";padding:2px 8px;border-radius:4px;font-size:12px;font-family:monospace;";
			kbd.textContent = items[i].key;
			row.appendChild(desc);
			row.appendChild(kbd);
			modal.appendChild(row);
		}
		overlay.appendChild(modal);
		document.body.appendChild(overlay);
		overlay.addEventListener("click", function(e) { if (e.target === overlay) overlay.remove(); });
	}, "Show keyboard shortcuts");
})();
</script>`
}

// RegisterShortcut renders a script that registers a single keyboard shortcut.
func RegisterShortcut(key, jsHandler, description string) string {
	return fmt.Sprintf(`<script>
if (window.__shortcuts) window.__shortcuts.register(%q, function(e){ %s }, %q);
</script>`, key, jsHandler, description)
}
