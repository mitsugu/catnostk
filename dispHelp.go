package main

import (
	"fmt"
	"os"
)

/*
dispHelp {{{
*/
func dispHelp() {
	usageTxt := `Usage :
	catnostk < <source file> > <destnation file>`
	fmt.Fprintf(os.Stderr, "%s\n", usageTxt)
}

// }}}
