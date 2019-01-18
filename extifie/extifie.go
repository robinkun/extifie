package extifie

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/robinkun/extifie/extifie/atominfo"
)

const (
	// HartreeKcalPerMol : 1 hartree = 627.5095 kcal / mol
	HartreeKcalPerMol = 627.5095
)

// Connectivity : CPFの中に書いてある結合情報を保存するための構造体
type Connectivity struct {
	src  int // 結合元
	dest int // 結合先
}

// FmoInfo : FMO結果を格納する構造体
type FmoInfo struct {
	CPFVersion             string
	ResidueName            map[int]string
	FragmentNum            int
	AtomNum                int
	Atom                   []atominfo.AtomInfo
	ConnectivityNum        int
	ConnectivityInfo       []Connectivity
	NuclearRepulsionEnergy [][]float64 // 核間反発エネルギー
	HfElectronEnergy       [][]float64 // HF-IFIEの電子エネルギー項
	HfElectroStaticEnergy  [][]float64 // HF-IFIEの静電エネルギー項
	Mp2Ifie                [][]float64
	HfIfieBsse             [][]float64
	Mp2IfieBsse            [][]float64
	Ifie                   [][]float64 // csvに出力するifie
}

// scanErrPanic : スキャナがエラーを吐いていたらプログラムを修正させる
func scanErrPanic(scanner *bufio.Scanner, errorMessage string) {
	if err := scanner.Err(); err != nil {
		fmt.Printf("[ERROR] %s", errorMessage)
		panic(err)
	}
}

// mallocFmoInfo : ConnectivityInfo&ResidueName以外のFmoInfo構造体のスライスを初期化しておく。
func mallocFmoInfo(fmoInfo *FmoInfo) {
	if fmoInfo.FragmentNum < 1 || fmoInfo.AtomNum < 1 {
		panic("[ERROR] Number of fragments or number of atoms are illegal.")
	}
	fmoInfo.Atom = make([]atominfo.AtomInfo, fmoInfo.AtomNum)

	fmoInfo.NuclearRepulsionEnergy = make([][]float64, fmoInfo.FragmentNum)
	fmoInfo.HfElectronEnergy = make([][]float64, fmoInfo.FragmentNum)
	fmoInfo.HfElectroStaticEnergy = make([][]float64, fmoInfo.FragmentNum)
	fmoInfo.Mp2Ifie = make([][]float64, fmoInfo.FragmentNum)
	fmoInfo.HfIfieBsse = make([][]float64, fmoInfo.FragmentNum)
	fmoInfo.Mp2IfieBsse = make([][]float64, fmoInfo.FragmentNum)
	fmoInfo.Ifie = make([][]float64, fmoInfo.FragmentNum)
	for i := 0; i < fmoInfo.FragmentNum; i++ {
		fmoInfo.NuclearRepulsionEnergy[i] = make([]float64, fmoInfo.FragmentNum)
		fmoInfo.HfElectronEnergy[i] = make([]float64, fmoInfo.FragmentNum)
		fmoInfo.HfElectroStaticEnergy[i] = make([]float64, fmoInfo.FragmentNum)
		fmoInfo.Mp2Ifie[i] = make([]float64, fmoInfo.FragmentNum)
		fmoInfo.HfIfieBsse[i] = make([]float64, fmoInfo.FragmentNum)
		fmoInfo.Mp2IfieBsse[i] = make([]float64, fmoInfo.FragmentNum)
		fmoInfo.Ifie[i] = make([]float64, fmoInfo.FragmentNum)
	}
}

// LoadCPF : CPFからデータを読み込んでFmoInfo型に格納
func (fmoInfo *FmoInfo) LoadCPF(path string) bool {
	fp, err := os.Open(path)

	if err != nil {
		panic(err)
	}
	// 関数終了時に確実にファイルが閉じるようにする
	defer fp.Close()

	scanner := bufio.NewScanner(fp)
	getGeom(fmoInfo, scanner)

	// MP2の文字列が現れるまで読み飛ばす
	for scanner.Scan() {
		scanErrPanic(scanner, "")
		if strings.Contains(scanner.Text(), "MP2") {
			break
		}
	}

	scanner.Scan() // 近似パラメータ
	scanner.Scan() // 核間反発エネルギー
	scanner.Scan() // 全電子エネルギー
	scanner.Scan() // 全エネルギー
	scanErrPanic(scanner, "")

	// モノマーのエネルギー等を読み飛ばす
	for i := 0; i < fmoInfo.FragmentNum; i++ {
		scanner.Scan()
		scanErrPanic(scanner, "")
	}

	// IFIE読み取り
	for i := 1; i < fmoInfo.FragmentNum; i++ {
		for j := 0; j < i; j++ {
			scanner.Scan()
			scanErrPanic(scanner, "")
			strs := strings.Fields(scanner.Text())
			if len(strs) != 18 {
				panic("[ERROR] File format error.")
			}
			fmoInfo.NuclearRepulsionEnergy[i][j], _ = strconv.ParseFloat(strs[0], 64)
			fmoInfo.HfElectronEnergy[i][j], _ = strconv.ParseFloat(strs[1], 64)
			fmoInfo.HfElectroStaticEnergy[i][j], _ = strconv.ParseFloat(strs[2], 64)
			fmoInfo.Mp2Ifie[i][j], _ = strconv.ParseFloat(strs[3], 64)
			fmoInfo.HfIfieBsse[i][j], _ = strconv.ParseFloat(strs[8], 64)
			fmoInfo.Mp2IfieBsse[i][j], _ = strconv.ParseFloat(strs[9], 64)

			fmoInfo.NuclearRepulsionEnergy[j][i] = fmoInfo.NuclearRepulsionEnergy[i][j]
			fmoInfo.HfElectronEnergy[j][i] = fmoInfo.HfElectronEnergy[i][j]
			fmoInfo.HfElectroStaticEnergy[j][i] = fmoInfo.HfElectroStaticEnergy[i][j]
			fmoInfo.Mp2Ifie[j][i] = fmoInfo.Mp2Ifie[i][j]
			fmoInfo.HfIfieBsse[j][i] = fmoInfo.HfIfieBsse[i][j]
			fmoInfo.Mp2IfieBsse[j][i] = fmoInfo.Mp2IfieBsse[i][j]

			// 隣り合っていたら0にするってやつを書く。
			fmoInfo.Ifie[i][j] = fmoInfo.NuclearRepulsionEnergy[i][j] + fmoInfo.HfElectronEnergy[i][j] + fmoInfo.Mp2Ifie[i][j]
			fmoInfo.Ifie[j][i] = fmoInfo.Ifie[i][j]
		}
	}

	for i := 0; i < fmoInfo.FragmentNum; i++ {
		for j := 0; j < fmoInfo.FragmentNum; j++ {
			fmt.Printf("%14f", fmoInfo.Ifie[i][j]*627.5095)
		}
		fmt.Println()
	}
	fmt.Println("Successful Loaded.")
	return true
}

func getGeom(fmoInfo *FmoInfo, scanner *bufio.Scanner) {
	// CPFバージョン読み取り
	scanner.Scan()
	scanErrPanic(scanner, "File is empty.")
	fmoInfo.CPFVersion = scanner.Text()

	// 原子数、フラグメント数読み取り
	scanner.Scan()
	scanErrPanic(scanner, "")
	// スペース区切りで文字列を読む
	strs := strings.Fields(scanner.Text())
	if len(strs) != 2 {
		panic("[ERROR] File format error.")
	}
	fmoInfo.AtomNum, _ = strconv.Atoi(strs[0])
	fmoInfo.FragmentNum, _ = strconv.Atoi(strs[1])

	mallocFmoInfo(fmoInfo)

	fmoInfo.ResidueName = make(map[int]string)

	// 原子情報読み取り
	//residueCnt := 0
	for i := 0; i < fmoInfo.AtomNum; i++ {
		scanner.Scan()
		scanErrPanic(scanner, "")
		// 原子情報が書いてある部分は15行だと決まっている
		strs = strings.Fields(scanner.Text())
		if len(strs) != 15 {
			panic("[ERROR] File format error.")
		}
		fmoInfo.Atom[i].AtomNum, _ = strconv.Atoi(strs[0])
		fmoInfo.Atom[i].AtomName = strs[1]
		fmoInfo.Atom[i].AtomType = strs[2]
		fmoInfo.Atom[i].ResidueNum, _ = strconv.Atoi(strs[4])
		fmoInfo.ResidueName[fmoInfo.Atom[i].ResidueNum] = strs[3]
		fmoInfo.Atom[i].FragmentNum, _ = strconv.Atoi(strs[5])
		fmoInfo.Atom[i].PrintAtomInfo()
	}

	// 電子数読み飛ばし
	for i := 0; i < (fmoInfo.FragmentNum-1)/16+1; i++ {
		scanner.Scan()
		scanErrPanic(scanner, "")
	}
	// 結合数読み取り
	fmoInfo.ConnectivityNum = 0
	for i := 0; i < (fmoInfo.FragmentNum-1)/16+1; i++ {
		scanner.Scan()
		scanErrPanic(scanner, "")
		strs = strings.Fields(scanner.Text())
		for j := 0; j < len(strs); j++ {
			tmp, _ := strconv.Atoi(strs[j])
			fmoInfo.ConnectivityNum += tmp
		}
	}
	//fmt.Println(fmoInfo.ConnectivityNum)

	// 結合情報を読み取り、フラグメント間の結合情報に変換
	fmoInfo.ConnectivityInfo = make([]Connectivity, fmoInfo.ConnectivityNum)
	for i := 0; i < fmoInfo.ConnectivityNum; i++ {
		scanner.Scan()
		scanErrPanic(scanner, "")
		strs = strings.Fields(scanner.Text())
		if len(strs) != 2 {
			panic("[ERROR] File format error.")
		}
		fmoInfo.ConnectivityInfo[i].src, _ = strconv.Atoi(strs[0])
		fmoInfo.ConnectivityInfo[i].dest, _ = strconv.Atoi(strs[1])
		//fmoInfo.ConnectivityInfo[i].src = fmoInfo.Atom[src-1].FragmentNum
		//fmoInfo.ConnectivityInfo[i].dest = fmoInfo.Atom[dest-1].FragmentNum
		fmt.Printf("%d, %d\n", fmoInfo.ConnectivityInfo[i].src, fmoInfo.ConnectivityInfo[i].dest)
	}
	/*
		// テストコード
		fmt.Println(fragmentCnt)
		for i := 0; i < fmoInfo.FragmentNum; i++ {
			fmt.Println(fmoInfo.ResidueName[i])
		}
	*/
}
