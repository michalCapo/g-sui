package ui

import (
	"cmp"
	"encoding/json"
	"net/http"
)

// Texts are the words the client shows by itself. Set App.Texts to show
// them in another language; empty fields keep English.
type Texts struct {
	Connecting     string `json:"connecting"`     // badge while reconnecting
	Offline        string `json:"offline"`        // badge when the connection is gone
	Reconnecting   string `json:"reconnecting"`   // next to Offline
	Loading        string `json:"loading"`        // loading badge, skeleton label
	PleaseWait     string `json:"pleaseWait"`     // next to Loading
	LostMaybeDone  string `json:"lostMaybeDone"`  // an action was sent, the answer was lost
	LostNotSent    string `json:"lostNotSent"`    // an action could not be sent
	OfflineNotSent string `json:"offlineNotSent"` // an action was tried while offline
	Dismiss        string `json:"dismiss"`        // label of a notification's close button
	Reload         string `json:"reload"`         // button that loads the page again
	ActionFailed   string `json:"actionFailed"`   // an action returned an error
	ServerError    string `json:"serverError"`    // an action panicked
	Denied         string `json:"denied"`         // an action was refused or its page is gone
}

func (t Texts) withDefaults() Texts {
	return Texts{
		Connecting:     cmp.Or(t.Connecting, "Connecting…"),
		Offline:        cmp.Or(t.Offline, "Offline"),
		Reconnecting:   cmp.Or(t.Reconnecting, "Trying to reconnect…"),
		Loading:        cmp.Or(t.Loading, "Loading…"),
		PleaseWait:     cmp.Or(t.PleaseWait, "Please wait"),
		LostMaybeDone:  cmp.Or(t.LostMaybeDone, "Connection lost. Your last action may have completed. Check before retrying."),
		LostNotSent:    cmp.Or(t.LostNotSent, "Connection lost before sending. Please try again."),
		OfflineNotSent: cmp.Or(t.OfflineNotSent, "You are offline. Reconnect before trying again."),
		Dismiss:        cmp.Or(t.Dismiss, "Dismiss notification"),
		Reload:         cmp.Or(t.Reload, "Reload"),
		ActionFailed:   cmp.Or(t.ActionFailed, "Unable to complete the request"),
		ServerError:    cmp.Or(t.ServerError, "Server error"),
		Denied:         cmp.Or(t.Denied, "Request denied or page expired"),
	}
}

// texts are the app's texts for server messages, English without an app.
// With App.Locale they are the request's texts.
func (ctx *Context) texts() Texts {
	if ctx == nil || ctx.app == nil {
		return Texts{}.withDefaults()
	}
	_, texts := ctx.app.locale(ctx.Request)
	return texts
}

// locale is the language and the texts of r: App.Locale's answer over Lang
// and Texts, with English for what is still empty.
func (app *App) locale(r *http.Request) (string, Texts) {
	lang, texts := app.Lang, app.Texts
	if app.Locale != nil && r != nil {
		l, t := app.Locale(r)
		lang = cmp.Or(l, lang)
		texts = t.or(texts)
	}
	return cmp.Or(lang, "en"), texts.withDefaults()
}

// or fills t's empty fields from other.
func (t Texts) or(other Texts) Texts {
	return Texts{
		Connecting:     cmp.Or(t.Connecting, other.Connecting),
		Offline:        cmp.Or(t.Offline, other.Offline),
		Reconnecting:   cmp.Or(t.Reconnecting, other.Reconnecting),
		Loading:        cmp.Or(t.Loading, other.Loading),
		PleaseWait:     cmp.Or(t.PleaseWait, other.PleaseWait),
		LostMaybeDone:  cmp.Or(t.LostMaybeDone, other.LostMaybeDone),
		LostNotSent:    cmp.Or(t.LostNotSent, other.LostNotSent),
		OfflineNotSent: cmp.Or(t.OfflineNotSent, other.OfflineNotSent),
		Dismiss:        cmp.Or(t.Dismiss, other.Dismiss),
		Reload:         cmp.Or(t.Reload, other.Reload),
		ActionFailed:   cmp.Or(t.ActionFailed, other.ActionFailed),
		ServerError:    cmp.Or(t.ServerError, other.ServerError),
		Denied:         cmp.Or(t.Denied, other.Denied),
	}
}

// json is the texts with defaults as a JS object literal. json.Marshal
// escapes <, > and &, so it is safe inside <script>.
func (t Texts) json() string {
	data, _ := json.Marshal(t.withDefaults())
	return string(data)
}

// js sets the texts for the client; it runs in <head> before the client.
func (t Texts) js() string {
	return "window.__gsuiText=" + t.json() + ";"
}
