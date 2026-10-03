# Building UI with g-sui

Rules for agents that write applications with this library. Full API:
[`docs/documentation.md`](docs/documentation.md). Working examples: `example/` (local actions: `example/pages/actions.go`).

## Rules

1. Build every element with Go nodes (`ui.Div`, `ui.Button`, ...). Never write
   HTML strings or `<script>` tags.
2. Use g-sui actions for behavior. Do not write JavaScript. `UnsafeJS`,
   `node.UnsafeJS` and `ctx.UnsafeHeadJS` are a last resort for trusted code
   that no API below covers. Never put user input in them.
3. Third-party JS libraries (charts, maps, editors) go in `app.Widget`.
4. UI state the server does not need (menus, tabs, dialogs, toggles) uses
   local actions. Data changes use server actions that return a `Result`.
5. Register a server action once, keep its `ActionRef` in a package variable,
   and call it from render code. Register actions before the pages that use them.
6. Style with Tailwind classes and `dark:` variants.

## Server action pattern

```go
type SaveInput struct {
    ID   int
    Name string
}

var saveItem ui.ActionRef[SaveInput]

func Register(app *ui.App) {
    saveItem = ui.RegisterAction(app, "item.save", func(ctx *ui.Context, in SaveInput) (ui.Result, error) {
        // save in.Name ...
        return ui.Result{}.SetText("status", "Saved").Toast("Saved"), nil
    })
    app.Page("/items", itemsPage)
}

ui.IText("...").ID("name")
ui.Button("...").Text("Save").OnClick(saveItem.Call(SaveInput{ID: 7}).Collect("name"))
```

Collected values fill input fields by name. A `Result` can also `Morph`,
`Replace`, `Append`, `Prepend`, `Remove`, `Refresh`, `Navigate` and `Run`
local actions.

## Want X? Use Y

| Want | Use |
| --- | --- |
| Show / hide / toggle a panel or menu | `ui.Show(id)`, `ui.Hide(id)`, `ui.Toggle(id)` |
| Close a menu on outside click | `menu.OnOutsideClick(ui.Hide(id))` |
| Tabs, active state, highlight | `ui.AddClass`, `ui.RemoveClass`, `ui.ToggleClass` in `ui.Seq(...)`, or `ui.NewTabs()` |
| Disable a button, set aria/data attributes | `ui.SetAttr`, `ui.RemoveAttr`, `ui.ToggleAttr` |
| Modal dialog | `ui.El("dialog")` with `ui.OpenDialog(id)` / `ui.CloseDialog(id)` |
| Confirm before an action | `ui.Confirm("Sure?", action)` or `ui.ConfirmDialog(...)` from a server action |
| Several steps on one click | `ui.Seq(a, b, c)` |
| Combine updates from several helpers | `ui.Merge(a, b)` / `r.Add(...)` |
| Run later | `ui.Delay(2*time.Second, action)` |
| Toast | `ui.Toast(msg)`, `ui.Notify(variant, msg)`; in a handler `Result.Toast` |
| Copy to clipboard | `ui.CopyText(text)`, `ui.CopyFrom(inputID)` |
| Set text or input value | `ui.SetText(id, text)`, `ui.SetValue(id, value)` |
| Clear, reset or submit a form | `ui.SetValue(id, "")`, `ui.ResetForm(id)`, `ui.SubmitForm(id)` |
| Focus or scroll | `ui.Focus(id)`, `ui.ScrollTo(id)`, `ui.ScrollTop()` |
| Remove an element | `ui.Remove(id)` |
| Show password | `ui.TogglePassword(id)` |
| Enter / Escape in an input | `input.OnKey("Enter", a).OnKey("Escape", b)` |
| Keyboard shortcut | `node.Shortcut("mod+k", ui.Focus("search"))` |
| Search as you type | `input.OnInput(search.Call(Q{}).Collect("q"))` (debounced) |
| React to select / checkbox | `node.OnChange(action)` |
| Links between pages | `ui.NavLink(path)`; in actions `ui.Navigate(path)` |
| Active nav link | `ui.NavLink(path).ActiveClass("active classes", "inactive classes")` |
| Change query, keep the view | `ui.PatchURL(path, replace)` |
| Full redirect, reload, back, print | `ui.Redirect(url)`, `ui.Reload()`, `ui.Back()`, `ui.Print()` |
| Page title | `ui.SetTitle(title)` |
| Dark mode | `ui.ThemeSwitcher()`, `ui.ToggleTheme()`, `ui.SetTheme("dark")` |
| Drag to scroll a wide table | `node.DragToScroll()` |
| Forms with validation | `ui.FormFor[T]` or `ui.NewForm(id).Action(ref)` |
| Data tables, lists with filters | `ui.RegisterTable`, `ui.NewDataTable[T]`, `ui.NewCollate[T]` |
| Live updates from the server | `ctx.Push`, `app.Broadcast`, `Subscribe`, `app.Live` |
| Re-auth socket after cookie refresh | `await window.__ws.restart()` |
| Chart, map, editor | `app.Widget(name, mountJS)` and `ui.Widget(name, props)` |

## Example: dropdown menu without JavaScript

```go
ui.Div("relative").Render(
    ui.Button("px-3 py-2 border rounded").Text("Menu").OnClick(ui.Toggle("menu")),
    ui.Div("hidden absolute mt-1 bg-white dark:bg-gray-900 shadow rounded").ID("menu").
        OnOutsideClick(ui.Hide("menu")).
        OnKey("Escape", ui.Hide("menu")).
        Render(
            ui.Button("block px-3 py-2").Text("Copy link").OnClick(ui.Seq(ui.CopyText(url), ui.Hide("menu"))),
            ui.Button("block px-3 py-2 text-red-600").Text("Delete").
                OnClick(ui.Confirm("Delete item?", deleteItem.Call(DeleteInput{ID: id}))),
        ),
)
```
