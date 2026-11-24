package pages

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"strings"
	"unicode/utf8"

	r "github.com/michalCapo/g-sui/ui"
)

const (
	form1ID = "form1"
	form2ID = "form2"
)

type sharedInput struct {
	FormID      string `json:"formID"`
	Title       string `json:"Title"`
	Description string `json:"Description"`
}

func (input *sharedInput) Validate() error {
	if input.FormID != form1ID && input.FormID != form2ID {
		return errors.New("unknown form")
	}
	return nil
}

var sharedSubmit, sharedReset r.ActionRef[sharedInput]

func sharedForm(formID, title, description string) *r.Node {
	inputCls := "w-full border border-gray-300 rounded px-3 py-2 text-sm"

	return r.Div("flex flex-col gap-4").ID(formID).Render(
		r.Div("flex flex-col gap-1").Render(
			r.Div("text-gray-600 text-sm").Text("Title"),
			r.IText(inputCls).ID(formID+"-title").Attr("name", "Title").Attr("value", title).Attr("placeholder", "Title"),
		),
		r.Div("flex flex-col gap-1").Render(
			r.Div("text-gray-600 text-sm").Text("Description"),
			r.Textarea(inputCls).ID(formID+"-desc").Attr("name", "Description").Attr("placeholder", "Description").Text(description),
		),
		r.Div("flex flex-row gap-4 justify-end").Render(
			r.Button("rounded-lg hover:text-red-700 hover:underline text-gray-400 px-3 py-1 cursor-pointer text-sm").
				Text("Reset").
				OnClick(sharedReset.Call(sharedInput{FormID: formID})),
			r.Button("rounded-lg px-4 py-2 bg-blue-600 text-white cursor-pointer hover:bg-blue-700 text-sm").
				Text("Submit").
				OnClick(sharedSubmit.Call(sharedInput{FormID: formID}).Collect(formID+"-title", formID+"-desc")),
		),
	)
}

func Shared(_ *r.Context) *r.Node {
	return r.Div("max-w-5xl mx-auto flex flex-col gap-4").Render(
		r.Div("text-2xl font-bold").Text("Shared"),
		r.Div("text-gray-600").Text("Reused form template in multiple places."),

		r.Div("border rounded-lg p-4 bg-white shadow-lg").Render(
			r.Div("text-lg font-semibold").Text("Form 1"),
			r.Div("text-gray-600 text-sm mb-4").Text("This form is reused."),
			sharedForm(form1ID, "Hello", "What a nice day"),
		),
		r.Div("border rounded-lg p-4 bg-white shadow-lg").Render(
			r.Div("text-lg font-semibold").Text("Form 2"),
			r.Div("text-gray-600 text-sm mb-4").Text("This form is reused."),
			sharedForm(form2ID, "Next Title", "Next Description"),
		),
	)
}

func HandleSharedSubmit(_ *r.Context, data sharedInput) (r.Result, error) {
	formID := data.FormID
	title := data.Title
	desc := data.Description

	if formID == form1ID {
		return r.Result{}.
			Replace(formID, sharedForm(formID, title, desc)).
			Notify("error", "Data not stored"), nil
	}

	return r.Result{}.
		Replace(formID, sharedForm(formID, title, desc)).
		Notify("success", "Data stored"), nil
}

func HandleSharedReset(_ *r.Context, data sharedInput) (r.Result, error) {
	formID := data.FormID
	return r.Result{}.Replace(formID, sharedForm(formID, "", "")), nil
}

func RegisterShared(app *r.App) {
	sharedSubmit = r.RegisterAction(app, "shared.submit", HandleSharedSubmit)
	sharedReset = r.RegisterAction(app, "shared.reset", HandleSharedReset)
	app.Page("/shared", Shared)
}

// csvDownload returns rows as a CSV file download. encoding/csv quotes the
// cells, and csvCell stops spreadsheets from running a cell as a formula.
func csvDownload(filename string, rows [][]string) (r.Result, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, row := range rows {
		for i, cell := range row {
			row[i] = csvCell(cell)
		}
		if err := w.Write(row); err != nil {
			return r.Result{}, err
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return r.Result{}, err
	}
	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())
	return r.Result{}.Download(filename, "text/csv", b64), nil
}

// csvCell prefixes a cell that starts with = + - @ tab or CR with a single
// quote, so a spreadsheet shows it as text instead of running a formula.
func csvCell(cell string) string {
	// Spreadsheets also read full-width signs and leading line breaks.
	if first, _ := utf8.DecodeRuneInString(cell); strings.ContainsRune("=+-@\t\r\n＝＋－＠", first) {
		return "'" + cell
	}
	return cell
}
