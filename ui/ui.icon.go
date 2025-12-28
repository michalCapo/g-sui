package ui

// IconBasic is kept for backward compatibility
func IconBasic(class string, text string) string {
	return Div("flex-1 flex items-center gap-2")(
		Icon(class),
		Div("text-center")(text),
		Flex1,
	)
}
