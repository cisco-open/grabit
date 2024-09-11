package internal

// ANSI color codes.
var clear = "\033[0m"
var green = "\033[32"
var yellow = "\033[33"

/*
var red = "\033[31"
var blue = "\033[34"
var magenta = "\033[35"
var cyan = "\033[36"
var gray = "\033[37"
var white = "\033[97"
var bold_suffix = ";1m"
*/
var reg_suffix = "m"

// Use to get colored text for printing to terminal.
// Returns "<color_code> <str> <clear>"
// Ex: Color_Test("Hello", "green") returns "\033[32mHello\033[0m"
func Color_Text(str string, color string) string {
	var colStr string
	if color == "green" {
		colStr += green
	} else if color == "yellow" {
		colStr += yellow
	} else {
		//Invalid color provided.
		return str
	}

	colStr += reg_suffix + str + clear
	return colStr
}
