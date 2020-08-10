package log

import (
	"fmt"
)

// Log level for the GOclient
var level = 0

// Check if "app" has been executed with 1 argument only.
func Print(minLvl int, msgSlice ...string) {
	if level >= minLvl {
		for _, msg := range msgSlice {
			if minLvl != 0 {
				fmt.Printf("Debug %d: ", minLvl)
			}
			fmt.Print(msg)
		}
		fmt.Printf("\n")
	}
}

// Overwrite the logs level
func SetLevel(cfgLevel int) {
	level = cfgLevel
	msg := fmt.Sprintf("Setting log level to \"%d\"...", level)
	Print(0, msg)
}

// Overwrite the logs level
func GetLevel() int {
	return level
}
