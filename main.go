package main

import (
	"fmt"

	"github.com/robinkun/extifie/extifie"
)

func main() {
	filepath := "./testfiles/gly5_1.cpf"
	var fmoResult extifie.FmoResult
	fmoResult.LoadCPF(filepath)

	fmt.Println(filepath)
	extifie.Test()
}
