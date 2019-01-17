package extifie

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FmoResult : FMO結果を格納する構造体
type FmoResult struct {
	CPFVersion            string
	FragmentName          []string
	FragmentNum           int
	AtomNum               int
	HfBsseIfie            [][]float64
	Mp2bsseIfie           [][]float64
	Mp2Ifie               [][]float64
	HfElectronEnergy      [][]float64
	HfElectroStaticEnergy [][]float64
	Ifie                  [][]float64 // BSSEなど補正をした後のifieを入れる
}

// LoadCPF : CPFからデータを読み込んでFmoResult型に格納
func (fmoResult *FmoResult) LoadCPF(path string) bool {
	fp, err := os.Open(path)

	if err != nil {
		panic(err)
	}
	// 関数終了時に確実にファイルが閉じるようにする
	defer fp.Close()

	scanner := bufio.NewScanner(fp)

	// CPFバージョン読み取り
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		fmt.Println("[ERROR] File is empty.")
		panic(err)
	}

	fmoResult.CPFVersion = scanner.Text()

	// 原子数、フラグメント数読み取り
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	// スペース区切りで文字列を読む
	strs := strings.Fields(scanner.Text())
	if len(strs) != 2 {
		panic("[ERROR] File format error.")
	}
	fmoResult.AtomNum, _ = strconv.Atoi(strs[0])
	fmoResult.FragmentNum, _ = strconv.Atoi(strs[1])
	fmoResult.FragmentName = make([]string, fmoResult.FragmentNum)

	// フラグメント名読み取り
	fragmentCnt := 0
	for i := 0; i < fmoResult.AtomNum; i++ {
		scanner.Scan()
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		strs = strings.Fields(scanner.Text())
		if len(strs) != 15 {
			panic("[ERROR] File format error.")
		}
		fragmentNum, _ := strconv.Atoi(strs[5])
		if fragmentCnt < fragmentNum {
			fragmentCnt++
			if fragmentCnt > fmoResult.FragmentNum {
				panic("[ERROR] File format error.")
			}
			fmoResult.FragmentName[fragmentCnt-1] = strs[3]
		}
	}

	/*
		// テストコード
		fmt.Println(fragmentCnt)
		for i := 0; i < fmoResult.FragmentNum; i++ {
			fmt.Println(fmoResult.FragmentName[i])
		}
	*/

	// MP2の文字列が現れるまで読み飛ばす
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if strings.Contains(line, "MP2") {
			break
		}
	}

	fmt.Println("Successful Loaded.")
	return true
}

func Test() {
	fmt.Println("test")
}
