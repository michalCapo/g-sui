package ui

func Icon(class string, attr ...Attr) string {
	return Div(class, attr...)()
}
func IconBasic(class string, text string) string {
	return Div("flex-1 flex items-center gap-2")(
		Icon(class),
		Div("text-center")(text),
		Flex1,
	)
}

func IconStart(class string, text string) string {
	return Div("flex-1 flex items-center gap-2")(
		Icon(class),
		Flex1,
		Div("text-center")(text),
		Flex1,
	)
}

func IconLeft(class string, text string) string {
	return Div("flex-1 flex items-center gap-2")(
		Flex1,
		Icon(class),
		Div("text-center")(text),
		Flex1,
	)
}

func IconRight(class string, text string) string {
	return Div("flex-1 flex items-center gap-2")(
		Flex1,
		Div("text-center")(text),
		Icon(class),
		Flex1,
	)
}
func IconEnd(class string, text string) string {
	return Div("flex-1 flex items-center gap-2")(
		Flex1,
		Div("text-center")(text),
		Flex1,
		Icon(class),
	)
}
