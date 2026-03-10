package js

import "fmt"

// ContentSearch renders an in-page text search bar for a container.
// It uses TreeWalker to find text nodes and wraps matches in <mark> elements.
// containerSelector targets the content area to search within.
// triggerKey is the key that opens the search bar (default: "/").
func ContentSearch(containerSelector, triggerKey string) string {
	if triggerKey == "" {
		triggerKey = "/"
	}
	return fmt.Sprintf(`<script>
(function(){
	var trigger = %q;
	var containerSel = %q;
	var bar = null;
	var matchIdx = 0;
	var matchCount = 0;

	function isTyping() {
		var ae = document.activeElement;
		if (!ae) return false;
		var tag = ae.tagName;
		return tag === "INPUT" || tag === "TEXTAREA" || ae.isContentEditable;
	}

	function clearMarks(container) {
		var marks = container.querySelectorAll("mark[data-csearch]");
		for (var i = marks.length - 1; i >= 0; i--) {
			var m = marks[i];
			var parent = m.parentNode;
			parent.replaceChild(document.createTextNode(m.textContent), m);
			parent.normalize();
		}
	}

	function highlightMatches(container, query) {
		clearMarks(container);
		if (!query) { matchCount = 0; matchIdx = 0; return; }
		var walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, null, false);
		var nodes = [];
		while (walker.nextNode()) nodes.push(walker.currentNode);
		var q = query.toLowerCase();
		var count = 0;
		for (var i = 0; i < nodes.length; i++) {
			var node = nodes[i];
			var text = node.textContent;
			var lower = text.toLowerCase();
			var idx = lower.indexOf(q);
			if (idx < 0) continue;
			var frag = document.createDocumentFragment();
			var lastIdx = 0;
			while (idx >= 0) {
				if (idx > lastIdx) frag.appendChild(document.createTextNode(text.substring(lastIdx, idx)));
				var mark = document.createElement("mark");
				mark.setAttribute("data-csearch", String(count));
				mark.textContent = text.substring(idx, idx + query.length);
				mark.style.backgroundColor = "#fde68a";
				frag.appendChild(mark);
				count++;
				lastIdx = idx + query.length;
				idx = lower.indexOf(q, lastIdx);
			}
			if (lastIdx < text.length) frag.appendChild(document.createTextNode(text.substring(lastIdx)));
			node.parentNode.replaceChild(frag, node);
		}
		matchCount = count;
		matchIdx = 0;
		scrollToMatch(container);
	}

	function scrollToMatch(container) {
		var marks = container.querySelectorAll("mark[data-csearch]");
		for (var i = 0; i < marks.length; i++) {
			marks[i].style.backgroundColor = (i === matchIdx) ? "#f59e0b" : "#fde68a";
		}
		if (marks[matchIdx]) marks[matchIdx].scrollIntoView({ behavior: "smooth", block: "center" });
	}

	function createBar() {
		bar = document.createElement("div");
		bar.style.cssText = "position:fixed;top:0;left:50%%;transform:translateX(-50%%);z-index:10000;background:#fff;border:1px solid #d1d5db;border-radius:0 0 8px 8px;padding:8px 12px;display:flex;align-items:center;gap:8px;box-shadow:0 4px 12px rgba(0,0,0,0.1);";
		if (document.documentElement.classList.contains("dark")) {
			bar.style.background = "#1f2937";
			bar.style.borderColor = "#374151";
			bar.style.color = "#e5e7eb";
		}
		var inp = document.createElement("input");
		inp.type = "text";
		inp.placeholder = "Search in page...";
		inp.style.cssText = "border:1px solid #d1d5db;border-radius:4px;padding:4px 8px;font-size:14px;width:240px;outline:none;";
		if (document.documentElement.classList.contains("dark")) {
			inp.style.background = "#111827";
			inp.style.borderColor = "#4b5563";
			inp.style.color = "#e5e7eb";
		}
		var info = document.createElement("span");
		info.style.cssText = "font-size:12px;color:#6b7280;min-width:50px;text-align:center;";
		var closeBtn = document.createElement("button");
		closeBtn.textContent = "\u2715";
		closeBtn.style.cssText = "border:none;background:none;cursor:pointer;font-size:16px;color:#6b7280;padding:2px 6px;";
		closeBtn.onclick = function() { closeBar(); };

		inp.addEventListener("input", function() {
			var container = document.querySelector(containerSel);
			if (!container) return;
			highlightMatches(container, inp.value);
			info.textContent = matchCount > 0 ? (matchIdx + 1) + "/" + matchCount : "0/0";
		});
		inp.addEventListener("keydown", function(e) {
			if (e.key === "Escape") { closeBar(); return; }
			if (e.key === "Enter") {
				if (e.shiftKey) { matchIdx = (matchIdx - 1 + matchCount) %% matchCount; }
				else { matchIdx = (matchIdx + 1) %% matchCount; }
				var container = document.querySelector(containerSel);
				if (container) scrollToMatch(container);
				info.textContent = matchCount > 0 ? (matchIdx + 1) + "/" + matchCount : "0/0";
			}
		});

		bar.appendChild(inp);
		bar.appendChild(info);
		bar.appendChild(closeBtn);
		document.body.appendChild(bar);
		inp.focus();
	}

	function closeBar() {
		if (!bar) return;
		var container = document.querySelector(containerSel);
		if (container) clearMarks(container);
		bar.remove();
		bar = null;
		matchIdx = 0;
		matchCount = 0;
	}

	document.addEventListener("keydown", function(e) {
		if (e.key === trigger && !isTyping()) {
			e.preventDefault();
			if (bar) { closeBar(); } else { createBar(); }
		}
		if (e.key === "Escape" && bar) { closeBar(); }
	});
})();
</script>`, triggerKey, containerSelector)
}
