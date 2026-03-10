package js

import "fmt"

// AjaxForm renders a script that intercepts form submission and uses fetch instead.
// formID is the HTML ID of the form to intercept.
// opts can include: successRedirect, onSuccess (JS), onError (JS).
func AjaxForm(formID string, opts map[string]string) string {
	successRedirect := ""
	onSuccess := ""
	onError := ""
	if opts != nil {
		successRedirect = opts["successRedirect"]
		onSuccess = opts["onSuccess"]
		onError = opts["onError"]
	}

	return fmt.Sprintf(`<script>
(function(){
	var form = document.getElementById(%q);
	if (!form) return;
	form.addEventListener("submit", function(e) {
		e.preventDefault();
		var submitBtn = form.querySelector('button[type="submit"], input[type="submit"]');
		var origText = "";
		if (submitBtn) {
			origText = submitBtn.textContent || submitBtn.value;
			submitBtn.disabled = true;
			if (submitBtn.textContent !== undefined) submitBtn.textContent = "Loading...";
			else submitBtn.value = "Loading...";
			submitBtn.style.opacity = "0.6";
		}

		var fd = new FormData(form);
		var body = {};
		fd.forEach(function(v, k) { body[k] = v; });

		fetch(form.action || window.location.href, {
			method: form.method || "POST",
			headers: { "Content-Type": "application/json", "Accept": "application/json" },
			credentials: "same-origin",
			body: JSON.stringify(body)
		})
		.then(function(resp) {
			if (!resp.ok) throw new Error("HTTP " + resp.status);
			return resp.json();
		})
		.then(function(data) {
			if (submitBtn) {
				submitBtn.disabled = false;
				if (submitBtn.textContent !== undefined) submitBtn.textContent = origText;
				else submitBtn.value = origText;
				submitBtn.style.opacity = "1";
			}
			var redirect = %q;
			if (redirect) { window.location.href = redirect; return; }
			var onSuccessFn = %q;
			if (onSuccessFn) { (new Function("data", onSuccessFn))(data); return; }
			if (typeof __notify === "function") __notify(data.message || "Success", "success");
		})
		.catch(function(err) {
			if (submitBtn) {
				submitBtn.disabled = false;
				if (submitBtn.textContent !== undefined) submitBtn.textContent = origText;
				else submitBtn.value = origText;
				submitBtn.style.opacity = "1";
			}
			var onErrorFn = %q;
			if (onErrorFn) { (new Function("err", onErrorFn))(err); return; }
			if (typeof __notify === "function") __notify(err.message || "Request failed", "error");
			else alert(err.message || "Request failed");
		});
	});
})();
</script>`, formID, successRedirect, onSuccess, onError)
}
