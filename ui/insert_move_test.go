package ui

import (
	"os/exec"
	"strings"
	"testing"
)

// listDOMShim is a tiny DOM: enough for Insert, Move and Remove on a list.
const listDOMShim = `
function El(tag){this.tagName=tag;this.children=[];this.parentNode=null;this.id='';}
Object.defineProperty(El.prototype,'nextSibling',{get(){var s=this.parentNode.children;return s[s.indexOf(this)+1]||null}});
Object.defineProperty(El.prototype,'firstChild',{get(){return this.children[0]||null}});
El.prototype.remove=function(){if(!this.parentNode)return;var s=this.parentNode.children;s.splice(s.indexOf(this),1);this.parentNode=null};
El.prototype.prepend=function(c){c.remove();c.parentNode=this;this.children.unshift(c)};
El.prototype.appendChild=function(c){c.remove();c.parentNode=this;this.children.push(c);return c};
El.prototype.after=function(c){c.remove();var s=this.parentNode.children;c.parentNode=this.parentNode;s.splice(s.indexOf(this)+1,0,c)};
El.prototype.addEventListener=function(){};El.prototype.setAttribute=function(){};
var root=new El('body');
var document={createElement:function(t){return new El(t)},getElementById:function(id){var f=function(e){for(var c of e.children){if(c.id===id)return c;var r=f(c);if(r)return r}return null};return f(root)}};
var window={};var __gsui={remove:function(el,id){var e=document.getElementById(id);if(e)e.remove()}};
function list(ids){var l=new El('div');l.id='list';ids.forEach(function(id){var c=new El('div');c.id=id;c.textContent='old';l.appendChild(c)});root.appendChild(l);return l}
function order(){var l=document.getElementById('list');return l?l.children.map(function(c){return c.id+(c.textContent==='new'?'*':'')}).join(','):'none'}
`

func runListJS(t *testing.T, setup string, result Result) string {
	t.Helper()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not installed")
	}
	js, err := result.build(nil)
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(node, "-e", listDOMShim+setup+";"+js+";console.log(order())").CombinedOutput()
	if err != nil {
		t.Fatalf("node: %v\n%s", err, out)
	}
	return strings.TrimSpace(string(out))
}

func TestInsertPlacesNodeAfterAnchor(t *testing.T) {
	for _, tc := range []struct {
		name, setup string
		result      Result
		want        string
	}{
		{"first", "list(['a','b'])", Result{}.Insert("list", "", Div().ID("n").Text("new")), "n*,a,b"},
		{"after anchor", "list(['a','b'])", Result{}.Insert("list", "a", Div().ID("n").Text("new")), "a,n*,b"},
		{"after last", "list(['a','b'])", Result{}.Insert("list", "b", Div().ID("n").Text("new")), "a,b,n*"},
		{"missing anchor goes first", "list(['a'])", Result{}.Insert("list", "gone", Div().ID("n").Text("new")), "n*,a"},
		{"existing id is replaced and moved", "list(['a','b','c'])", Result{}.Insert("list", "c", Div().ID("a").Text("new")), "b,c,a*"},
		{"missing parent is skipped", "", Result{}.Insert("list", "", Div().ID("n")), "none"},
		{"move to first", "list(['a','b','c'])", Result{}.Move("c", ""), "c,a,b"},
		{"move after", "list(['a','b','c'])", Result{}.Move("a", "b"), "b,a,c"},
		{"move in place", "list(['a','b'])", Result{}.Move("b", "a"), "a,b"},
		{"move missing", "list(['a'])", Result{}.Move("x", "a"), "a"},
		{"list follows server order", "list(['a','b','c','d'])",
			Result{}.Remove("b").Insert("list", "", Div().ID("n").Text("new")).Move("d", "n").Move("a", "d"), "n*,d,a,c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := runListJS(t, tc.setup, tc.result); got != tc.want {
				t.Fatalf("order = %s, want %s", got, tc.want)
			}
		})
	}
}
