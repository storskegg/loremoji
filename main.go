package main

import (
	"fmt"
	"os"

	"github.com/storskegg/randemoter/application/randemoter"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Recovered Panic: ", err)
			os.Exit(100)
		}
	}()

	if err := randemoter.Execute(); err != nil {
		fmt.Println(err)
	}
}
