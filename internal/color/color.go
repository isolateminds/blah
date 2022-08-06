package color

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/ttacon/chalk"
)

func PrintFatal(message any) {
	debug.PrintStack()
	fmt.Println()
	log.Fatalf("(%s) %s", chalk.Red.Color("Error"), chalk.White.Color(fmt.Sprintf("%v", message)))
}
func PrintBlue(message any) {
	log.Println(chalk.Blue.Color(fmt.Sprintf("%v", message)))
}
func PrintCyan(message any) {
	log.Println(chalk.Cyan.Color(fmt.Sprintf("%v", message)))
}

func PrintYellow(message any) {
	log.Println(chalk.Yellow.Color(fmt.Sprintf("%v", message)))
}

func PrintGreen(message any) {
	log.Println(chalk.Green.Color(fmt.Sprintf("%v", message)))
}
func PrintStatus(state string, message any) {
	fmt.Printf("(%s) %s\n", chalk.White.Color(state), chalk.Green.Color(fmt.Sprintf("%v", message)))
}
func PrintStatusWithNewLine(state string, message any) {
	fmt.Printf("(%s)\n%s\n", chalk.White.Color(state), chalk.Green.Color(fmt.Sprintf("%v", message)))
}
func PrintForInput(message string) {
	fmt.Print(chalk.White.Color(message))
}
