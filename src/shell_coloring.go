package main

import "fmt"

// type bashColor represents escaped color codes for bash shell
type bashColor string 

const (
   COLOR_BOLD bashColor = "\033[1m"
   COLOR_UNDERLINED bashColor = "\033[4m"
   COLOR_INVERTED bashColor = "\033[7m"
   COLOR_RESET_BOLD bashColor = "\033[21m"
   COLOR_RESET_UNDERLINED bashColor = "\033[24m"
   COLOR_RESET_INVERTED bashColor = "\033[27m"
   COLOR_RESET_ALL bashColor = "\033[0m"
   COLOR_BLUE bashColor = "\033[34m"
   COLOR_YELLOW bashColor = "\033[33m"
   COLOR_RED bashColor = "\033[31m"
   COLOR_GREEN bashColor = "\033[32m"
   COLOR_CYAN bashColor = "\033[36m"
   COLOR_MAGENTA bashColor = "\033[35m"
   COLOR_LIGHT_BLUE bashColor = "\033[94m"
   COLOR_LIGHT_GREEN bashColor = "\033[92m"
   COLOR_LIGHT_YELLOW bashColor = "\033[93m"
   COLOR_LIGHT_RED bashColor = "\033[91m"
)
// Styles string s using shell (bash) color-code color
func Style(s string, color bashColor) string {
	if s == "" {
		return ""
	} else if color == "" {
		return s
   }
   return fmt.Sprintf("%v%v%v", color, s, COLOR_RESET_ALL)
}

