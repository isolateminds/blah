package cmd

import (
	"fmt"

	"github.com/ttacon/chalk"

	"github.com/spf13/cobra"
)

func printBanner() {
	printBlue := func(m string) {
		fmt.Println(chalk.Blue.Color(m))
	}
	printCyan := func(m string) {
		fmt.Println(chalk.Cyan.Color(m))
	}
	fmt.Println()
	printBlue(`   ___  __   ___   __ __ `)
	printBlue(`  / _ )/ /  / _ | / // / `)
	printBlue(` / _  / /__/ __ |/ _  /  `)
	printBlue(`/____/____/_/ |_/_//_/   `)
	printBlue(`The quick mongodb nginx setup utility.`)
	fmt.Println()
	printCyan(`Try --help           © 2022 dannyvidal`)
	fmt.Println()
}

var (
	rootCmd = &cobra.Command{
		Use:                "blah",
		CompletionOptions:  cobra.CompletionOptions{DisableDefaultCmd: true},
		DisableSuggestions: true,
		Run: func(cmd *cobra.Command, args []string) {

			if len(args) == 0 {
				printBanner()
			} else {
				fmt.Println(args)
			}
		},
	}
)

func Execute() {
	rootCmd.Execute()
}
