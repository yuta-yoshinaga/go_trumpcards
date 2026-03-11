package presenter

import "strings"

// buildCuiOutput constructs a CUI output string with a standard "==========" header block and footer.
// The output format is:
//
//	==========
//	<title>
//	==========
//	<content from buildContent>
//	==========
func buildCuiOutput(title string, buildContent func(b *strings.Builder)) string {
	var b strings.Builder
	b.WriteString("==========\n")
	b.WriteString(title + "\n")
	b.WriteString("==========\n")
	buildContent(&b)
	b.WriteString("==========\n")
	return b.String()
}
