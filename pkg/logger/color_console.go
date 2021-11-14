package logger

const (
	suffixReset = "\033[0m"

	// font style
	// normal    = 0
	// bold      = 1
	// dim       = 2
	// underline = 4
	// blink     = 5
	// reverse   = 7
	// hidden    = 8

	// font color
	// black       = 30 // default = 39
	// red         = 31
	// green       = 32
	// yellow      = 33
	// blue        = 34
	// purple      = 35 // purple = magenta
	// cyan        = 36
	// lightGray   = 37
	// darkGray    = 90
	// lightRed    = 91
	// lightGreen  = 92
	// lightYellow = 93
	// lightBlue   = 94
	// lightPurple = 95
	// lightCyan   = 96
	// white       = 97

	prefixCyan    = "\033[0;36m"
	prefixLRed    = "\033[0;91m"
	prefixLGreen  = "\033[0;92m"
	prefixLYellow = "\033[0;93m"
	prefixBRed    = "\033[1;31m"
	prefixBPurble = "\033[1;35m"
)
