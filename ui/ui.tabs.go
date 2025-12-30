package ui

import (
	"fmt"
	"strings"
)

// Tab style constants
const (
	TabsStylePills     = "pills"
	TabsStyleUnderline = "underline"
	TabsStyleBoxed     = "boxed"
	TabsStyleVertical  = "vertical"
)

// tabData represents a single tab with its label and content
type tabData struct {
	label   string
	icon    string
	content string
}

// tabs provides tabbed navigation for content organization
type tabs struct {
	tabs    []tabData
	active  int
	style   string
	class   string
	visible bool
	id      string
}

// Tabs creates a new tabs component for tabbed navigation
func Tabs() *tabs {
	id := "tabs_" + RandomString(8)
	return &tabs{
		tabs:    make([]tabData, 0),
		active:  0,
		style:   TabsStyleUnderline,
		visible: true,
		id:      id,
	}
}

// Tab adds a new tab with the given label and content
func (t *tabs) Tab(label, content string, icon ...string) *tabs {
	iconStr := ""
	if len(icon) > 0 {
		iconStr = icon[0]
	}
	t.tabs = append(t.tabs, tabData{
		label:   label,
		icon:    iconStr,
		content: content,
	})
	return t
}

// Active sets the initially active tab by index (0-based)
func (t *tabs) Active(index int) *tabs {
	if index >= 0 && index < len(t.tabs) {
		t.active = index
	}
	return t
}

// Style sets the visual style of the tabs: "pills", "underline", "boxed", or "vertical"
func (t *tabs) Style(value string) *tabs {
	switch value {
	case TabsStylePills, TabsStyleUnderline, TabsStyleBoxed, TabsStyleVertical:
		t.style = value
	default:
		t.style = TabsStyleUnderline
	}
	return t
}

// If conditionally renders the tabs based on the boolean value
func (t *tabs) If(value bool) *tabs {
	t.visible = value
	return t
}

// Class adds custom CSS classes to the tabs container
func (t *tabs) Class(value ...string) *tabs {
	t.class = strings.Join(value, " ")
	return t
}

// Render generates the HTML for the tabs component with JavaScript for tab switching
func (t *tabs) Render() string {
	if !t.visible || len(t.tabs) == 0 {
		return ""
	}

	// Generate unique IDs for tab buttons and panels
	buttonIDs := make([]string, len(t.tabs))
	panelIDs := make([]string, len(t.tabs))
	for i := range t.tabs {
		suffix := RandomString(6)
		buttonIDs[i] = fmt.Sprintf("%s_btn_%d_%s", t.id, i, suffix)
		panelIDs[i] = fmt.Sprintf("%s_panel_%d_%s", t.id, i, suffix)
	}

	var builder strings.Builder

	// Add scrollbar-hide CSS
	builder.WriteString(`<style>.scrollbar-hide::-webkit-scrollbar{display:none}.scrollbar-hide{-ms-overflow-style:none;scrollbar-width:none}</style>`)

	// Container div
	isVertical := t.style == TabsStyleVertical
	containerClass := Classes(
		"w-full",
		If(isVertical, func() string { return "flex flex-col md:flex-row gap-6" }),
		t.getClass(),
	)
	builder.WriteString(fmt.Sprintf(`<div id="%s" class="%s" data-tabs-active="%d" data-tabs-style="%s">`,
		escapeAttr(t.id),
		escapeAttr(containerClass),
		t.active,
		t.style,
	))

	// Render tab buttons based on style
	builder.WriteString(t.renderTabButtons(buttonIDs, panelIDs))

	// Render tab panels container if vertical
	if isVertical {
		builder.WriteString(`<div class="flex-1">`)
	}

	// Render tab panels
	builder.WriteString(t.renderTabPanels(panelIDs))

	if isVertical {
		builder.WriteString(`</div>`)
	}

	builder.WriteString(`</div>`)

	// Render JavaScript for tab switching
	builder.WriteString(t.renderJavaScript(buttonIDs, panelIDs))

	return builder.String()
}

// getClass returns the base CSS classes based on the tabs style
func (t *tabs) getClass() string {
	if t.class != "" {
		return t.class
	}
	return ""
}

// renderTabButtons generates the HTML for tab buttons based on the selected style
func (t *tabs) renderTabButtons(buttonIDs []string, panelIDs []string) string {
	var builder strings.Builder

	wrapperClass := "flex overflow-x-auto scrollbar-hide "
	switch t.style {
	case TabsStylePills:
		wrapperClass += "gap-2 mb-4"
	case TabsStyleUnderline:
		wrapperClass += "border-b border-gray-200 dark:border-gray-800 mb-4"
	case TabsStyleBoxed:
		wrapperClass += "gap-0 mb-4 border border-gray-200 dark:border-gray-800 rounded-lg overflow-hidden p-1 bg-gray-50/50 dark:bg-gray-950/30"
	case TabsStyleVertical:
		wrapperClass = "flex flex-col gap-1 min-w-[12rem]"
	}

	builder.WriteString(fmt.Sprintf(`<div class="%s" role="tablist">`, wrapperClass))

	for i, tab := range t.tabs {
		isActive := i == t.active
		buttonClass := t.getButtonClass(isActive)
		ariaSelected := If(isActive, func() string { return "true" })
		ariaControls := panelIDs[i]
		tabIndex := Or(isActive, func() string { return "0" }, func() string { return "-1" })

		builder.WriteString(fmt.Sprintf(
			`<button id="%s" class="%s" data-tabs-index="%d" role="tab" aria-selected="%s" aria-controls="%s" tabindex="%s">`,
			escapeAttr(buttonIDs[i]),
			escapeAttr(buttonClass),
			i,
			escapeAttr(ariaSelected),
			escapeAttr(ariaControls),
			escapeAttr(tabIndex),
		))

		if tab.icon != "" {
			builder.WriteString(fmt.Sprintf(`<span class="mr-2">%s</span>`, tab.icon))
		}
		builder.WriteString(fmt.Sprintf(`<span>%s</span>`, tab.label))
		builder.WriteString(`</button>`)
	}

	builder.WriteString(`</div>`)
	return builder.String()
}

// getButtonClass returns CSS classes for a tab button based on style and active state
func (t *tabs) getButtonClass(isActive bool) string {
	baseClass := "cursor-pointer font-bold transition-all duration-200 focus:outline-none text-sm whitespace-nowrap flex items-center justify-center"

	switch t.style {
	case TabsStylePills:
		activeClass := "bg-blue-600 text-white shadow-md shadow-blue-500/20"
		inactiveClass := "bg-transparent text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-800/50"
		if isActive {
			return Classes(baseClass, activeClass, "rounded-lg px-4 py-2")
		}
		return Classes(baseClass, inactiveClass, "rounded-lg px-4 py-2")

	case TabsStyleUnderline:
		activeClass := "text-blue-600 border-b-2 border-blue-600 dark:text-blue-400 dark:border-blue-400"
		inactiveClass := "text-gray-500 border-b-2 border-transparent hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600"
		if isActive {
			return Classes(baseClass, activeClass, "px-4 py-2.5 -mb-px")
		}
		return Classes(baseClass, inactiveClass, "px-4 py-2.5 -mb-px")

	case TabsStyleBoxed:
		activeClass := "bg-white text-blue-600 shadow-sm dark:bg-gray-800 dark:text-blue-400 rounded-md"
		inactiveClass := "text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
		if isActive {
			return Classes(baseClass, activeClass, "px-4 py-1.5 flex-1")
		}
		return Classes(baseClass, inactiveClass, "px-4 py-1.5 flex-1")

	case TabsStyleVertical:
		activeClass := "bg-blue-50 text-blue-700 border-r-2 border-blue-600 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400"
		inactiveClass := "text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 border-r-2 border-transparent"
		if isActive {
			return Classes(baseClass, activeClass, "px-4 py-3 text-left rounded-l-md")
		}
		return Classes(baseClass, inactiveClass, "px-4 py-3 text-left rounded-l-md")

	default:
		return Classes(baseClass, "px-4 py-2")
	}
}

// renderTabPanels generates the HTML for all tab panels
func (t *tabs) renderTabPanels(panelIDs []string) string {
	var builder strings.Builder

	for i, tab := range t.tabs {
		isActive := i == t.active
		hiddenAttr := If(!isActive, func() string { return `hidden=""` })
		labelledBy := t.id + "_btn_" + fmt.Sprint(i)

		panelClass := Classes(
			"tab-panel",
			If(!isActive, func() string { return "hidden opacity-0" }),
			If(isActive, func() string { return "opacity-100" }),
			"transition-opacity duration-300 ease-in-out",
		)

		builder.WriteString(fmt.Sprintf(
			`<div id="%s" class="%s" data-tabs-panel="%d" role="tabpanel" aria-labelledby="%s" %s>`,
			escapeAttr(panelIDs[i]),
			escapeAttr(panelClass),
			i,
			escapeAttr(labelledBy),
			hiddenAttr,
		))
		builder.WriteString(tab.content)
		builder.WriteString(`</div>`)
	}

	return builder.String()
}

// renderJavaScript generates the JavaScript code for managing tab interactions
func (t *tabs) renderJavaScript(buttonIDs, panelIDs []string) string {
	// Build style-specific active/inactive class information
	var activeClasses, inactiveClasses string
	switch t.style {
	case TabsStylePills:
		activeClasses = "bg-blue-600 text-white shadow-md shadow-blue-500/20"
		inactiveClasses = "bg-transparent text-gray-500 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:bg-gray-800/50"
	case TabsStyleUnderline:
		activeClasses = "text-blue-600 border-b-2 border-blue-600 dark:text-blue-400 dark:border-blue-400"
		inactiveClasses = "text-gray-500 border-b-2 border-transparent hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-200 dark:hover:border-gray-600"
	case TabsStyleBoxed:
		activeClasses = "bg-white text-blue-600 shadow-sm dark:bg-gray-800 dark:text-blue-400 rounded-md"
		inactiveClasses = "text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
	case TabsStyleVertical:
		activeClasses = "bg-blue-50 text-blue-700 border-r-2 border-blue-600 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-400"
		inactiveClasses = "text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-gray-800 border-r-2 border-transparent"
	default:
		activeClasses = "text-blue-600"
		inactiveClasses = "text-gray-600"
	}

	script := fmt.Sprintf(`<script>(function(){
		var container=document.getElementById('%s');
		if(!container)return;
		var buttons=container.querySelectorAll('button[data-tabs-index]');
		var panels=container.querySelectorAll('div[data-tabs-panel]');
		var activeClasses='%s';
		var inactiveClasses='%s';

		function setActiveTab(index){
			buttons.forEach(function(btn){
				var idx=parseInt(btn.getAttribute('data-tabs-index'));
				if(idx===index){
					btn.setAttribute('aria-selected','true');
					inactiveClasses.split(' ').filter(c => c).forEach(c => btn.classList.remove(c));
					activeClasses.split(' ').filter(c => c).forEach(c => btn.classList.add(c));
					btn.tabIndex=0;
				}else{
					btn.setAttribute('aria-selected','false');
					activeClasses.split(' ').filter(c => c).forEach(c => btn.classList.remove(c));
					inactiveClasses.split(' ').filter(c => c).forEach(c => btn.classList.add(c));
					btn.tabIndex=-1;
				}
			});
			panels.forEach(function(panel){
				var idx=parseInt(panel.getAttribute('data-tabs-panel'));
				if(idx===index){
					panel.classList.remove('hidden');
					panel.removeAttribute('hidden');
					setTimeout(() => {
						panel.classList.remove('opacity-0');
						panel.classList.add('opacity-100');
					}, 10);
					panel.setAttribute('aria-hidden','false');
				}else{
					panel.classList.add('hidden', 'opacity-0');
					panel.setAttribute('hidden', '');
					panel.classList.remove('opacity-100');
					panel.setAttribute('aria-hidden','true');
				}
			});
			container.setAttribute('data-tabs-active',index);
		}

		buttons.forEach(function(btn){
			btn.addEventListener('click',function(){
				var index=parseInt(this.getAttribute('data-tabs-index'));
				setActiveTab(index);
			});
			btn.addEventListener('keydown',function(e){
				var currentIndex=parseInt(container.getAttribute('data-tabs-active'));
				if(e.key==='ArrowRight' || e.key==='ArrowDown'){
					var newIndex=(currentIndex+1)%%buttons.length;
					buttons[newIndex].focus();
					setActiveTab(newIndex);
					e.preventDefault();
				}else if(e.key==='ArrowLeft' || e.key==='ArrowUp'){
					var newIndex=(currentIndex-1+buttons.length)%%buttons.length;
					buttons[newIndex].focus();
					setActiveTab(newIndex);
					e.preventDefault();
				}
			});
		});

		setActiveTab(%d);
	})();</script>`,
		escapeAttr(t.id),
		escapeJS(activeClasses),
		escapeJS(inactiveClasses),
		t.active,
	)
	return script
}
