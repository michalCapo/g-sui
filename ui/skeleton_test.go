package ui

import (
	"os"
	"strings"
	"testing"
)

// The golden file holds the skeletons as they rendered before SkeletonStyle,
// plus the reduced-motion class. Defaults must not change.
func TestSkeletonDefaultsMatchGolden(t *testing.T) {
	golden, err := os.ReadFile("testdata/skeletons.golden")
	if err != nil {
		t.Fatal(err)
	}
	want := strings.Split(strings.TrimSpace(string(golden)), "\n")
	var s SkeletonStyle
	got := []struct {
		name      string
		pkg, zero *Node
	}{
		{"Table", SkeletonTable(), s.Table()},
		{"Cards", SkeletonCards(), s.Cards()},
		{"List", SkeletonList(), s.List()},
		{"Component", SkeletonComponent(), s.Component()},
		{"Page", SkeletonPage(), s.Page()},
		{"Form", SkeletonForm(), s.Form()},
	}
	if len(want) != len(got) {
		t.Fatalf("golden has %d lines, want %d", len(want), len(got))
	}
	for i, g := range got {
		if js := g.pkg.DebugJS(); js != want[i] {
			t.Errorf("Skeleton%s changed:\n got %s\nwant %s", g.name, js, want[i])
		}
		if js := g.zero.DebugJS(); js != want[i] {
			t.Errorf("SkeletonStyle{}.%s differs from Skeleton%s", g.name, g.name)
		}
	}
}

func allSkeletons(s SkeletonStyle) map[string]string {
	return map[string]string{
		"Table": s.Table().DebugJS(), "Cards": s.Cards().DebugJS(), "List": s.List().DebugJS(),
		"Component": s.Component().DebugJS(), "Page": s.Page().DebugJS(), "Form": s.Form().DebugJS(),
	}
}

func TestSkeletonStyleClasses(t *testing.T) {
	s := SkeletonStyle{
		BarClass:     "bg-zinc-200 dark:bg-zinc-800",
		RoundClass:   "rounded-md",
		CardClass:    "bg-transparent",
		DividerClass: "border-zinc-100 dark:border-zinc-900",
	}
	for name, js := range allSkeletons(s) {
		for _, old := range []string{"gray-", "bg-white"} {
			if strings.Contains(js, old) {
				t.Errorf("%s still uses %q", name, old)
			}
		}
		if !strings.Contains(js, "bg-zinc-200 dark:bg-zinc-800") || !strings.Contains(js, "animate-pulse motion-reduce:animate-none") {
			t.Errorf("%s misses bar or pulse classes: %s", name, js)
		}
		if !strings.Contains(js, "rounded-md") {
			t.Errorf("%s misses RoundClass", name)
		}
	}
	js := allSkeletons(s)
	if !strings.Contains(js["Cards"], "'bg-transparent p-5 flex flex-col gap-3'") {
		t.Errorf("CardClass not used: %s", js["Cards"])
	}
	for _, name := range []string{"Table", "List", "Page"} {
		if !strings.Contains(js[name], "border-zinc-100 dark:border-zinc-900") {
			t.Errorf("%s misses DividerClass", name)
		}
	}
}

func TestSkeletonStyleCounts(t *testing.T) {
	s := SkeletonStyle{NoAvatar: true, Rows: 8, Lines: 1}
	if js := s.List().DebugJS(); strings.Contains(js, "rounded-full") || strings.Count(js, "'flex items-center gap-4 py-3'") != 8 {
		t.Errorf("List with NoAvatar and 8 rows: %s", js)
	}
	if n := strings.Count(s.Table().DebugJS(), "createElement('tr')"); n != 9 {
		t.Errorf("Table rows = %d, want 8 + header", n)
	}
	if n := strings.Count(s.Cards().DebugJS(), "p-5 flex flex-col gap-3"); n != 8 {
		t.Errorf("Cards = %d, want 8", n)
	}
	if n := strings.Count(s.Form().DebugJS(), "h-10 bg-gray-200 dark:bg-gray-700 rounded w-full"); n != 8 {
		t.Errorf("Form fields = %d, want 8", n)
	}
	if n := strings.Count(s.Component().DebugJS(), "'h-3 "); n != 1 {
		t.Errorf("Component lines = %d, want 1", n)
	}
	// Negative counts fall back to one item instead of panicking.
	if js := (SkeletonStyle{Rows: -1}).List().DebugJS(); strings.Count(js, "py-3") != 1 {
		t.Errorf("Rows -1: %s", js)
	}
}

func TestSkeletonParts(t *testing.T) {
	js := SkeletonBlock("flex gap-2", SkeletonBar("h-4 w-1/2")).DebugJS()
	want := "e0.className='animate-pulse motion-reduce:animate-none flex gap-2';var e1=document.createElement('div');e1.className='h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/2';"
	if !strings.Contains(js, want) {
		t.Errorf("parts:\n got %s\nwant %s", js, want)
	}
	if js := (SkeletonStyle{BarClass: "bg-zinc-300"}).Bar("h-2").DebugJS(); !strings.Contains(js, "'h-2 bg-zinc-300 rounded'") {
		t.Errorf("styled bar: %s", js)
	}
}

func TestAppSkeletonConfig(t *testing.T) {
	app := NewApp()
	app.InstanceID = "i"
	app.Loader = LoaderNone
	app.SkeletonDelayMs = -1
	app.SkeletonTarget = "main-content"
	node := Div("x").Attr("aria-label", "Loading mail")
	app.Skeleton("/", SkeletonForm())
	app.Skeleton("/mail/", node)
	app.Skeleton("/mail/drafts", nil)
	app.Skeleton("/", SkeletonList()) // replaces the earlier "/"

	cfg := app.wsConfigJS()
	prefix := `window.__gsuiCfg={grace:12000,reloadAfter:15000,keepAlive:25000,inst:'i',loader:'none',skelDelay:0,skelTarget:'main-content',skeletons:[['/mail/drafts',null],['/mail',function(){`
	if !strings.HasPrefix(cfg, prefix) || !strings.HasSuffix(cfg, "]};") {
		t.Fatalf("cfg:\n%s", cfg)
	}
	if strings.Count(cfg, "['/',") != 1 || !strings.Contains(cfg, "flex items-center gap-4 py-3") || strings.Contains(cfg, "h-10 w-32 mt-2") {
		t.Errorf("re-registered prefix should replace the old skeleton: %s", cfg)
	}
	for _, want := range []string{"'role','status'", "'aria-label','Loading mail'", "'aria-label','Loading'", "'data-gsui-skeleton',''"} {
		if !strings.Contains(cfg, want) {
			t.Errorf("cfg misses %s", want)
		}
	}
	if len(node.attrs) != 1 {
		t.Errorf("Skeleton changed the caller's node: %v", node.attrs)
	}
}

func TestActionNoLoader(t *testing.T) {
	type in struct{ ID int }
	ref := ActionRef[in]{name: "item.save"}
	a := ref.Call(in{ID: 1})
	quiet := a.NoLoader().Collect("name")
	if got, want := quiet.code(), `__ws.call("item.save",{"ID":1},["name"],el,{loader:false});`; got != want {
		t.Errorf("quiet code = %s, want %s", got, want)
	}
	if got := a.code(); strings.Contains(got, "loader") {
		t.Errorf("NoLoader changed the original action: %s", got)
	}
	if (*Action)(nil).NoLoader() != nil {
		t.Error("nil action")
	}
}
