package main

import (
	"flag"

	"github.com/robinkun/extifie/extifie"
)

func main() {
	var (
		inputFilePath  = flag.String("f", "", "Path to input cpf file")
		outputFilePath = flag.String("o", "", "Path to output csv file")
	)
	flag.Parse()
	//filepath := "./testfiles/gly5_1.cpf"
	var fmoResult extifie.FmoInfo
	fmoResult.LoadCPF(*inputFilePath)
	fmoResult.SetUnitKcalPerMol()
	fmoResult.GenerateCSV(*outputFilePath)
}
