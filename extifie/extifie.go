package extifie

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/robinkun/extifie/extifie/atominfo"
)

const (
	// HartreeKcalPerMol : 1 hartree = 627.5095 kcal / mol
	hartreeKcalPerMol = 627.5095
)

const (
	// Hartree : 単位をHartreeにする
	Hartree = iota
	// KcalPerMol : 単位をKcal/molにする
	KcalPerMol
)

// Connectivity : CPFの中に書いてある結合情報を保存するための構造体
type Connectivity struct {
	src  int // 結合元
	dest int // 結合先
}

// FmoInfo : FMO結果を格納する構造体
type FmoInfo struct {
	CPFVersion             string              // CPFのバージョン
	ResidueName            map[int]string      // 残基名
	FragmentNum            int                 // フラグメント数
	ResidueInFragment      []map[int]struct{}  // フラグメントに含まれる残基の番号のリスト。キーを値として使う。
	AtomNum                int                 // 原子の総数
	Atom                   []atominfo.AtomInfo // 原子ごとの情報
	ConnectivityNum        int                 // 結合情報の数
	ConnectivityInfo       []Connectivity      // 原子間の結合情報
	ConnectivityMatrix     [][]bool            // フラグメント間の結合情報
	NuclearRepulsionEnergy [][]float64         // 核間反発エネルギー
	HfElectronEnergy       [][]float64         // HF-IFIEの電子エネルギー項
	HfElectroStaticEnergy  [][]float64         // HF-IFIEの静電エネルギー項
	Mp2Ifie                [][]float64
	HfIfieBsse             [][]float64
	Mp2IfieBsse            [][]float64
	Ifie                   [][]float64 // csvに出力するifie
	unit                   int         // 出力する単位(Hartree or Kcal/mol)
}

// scanErrPanic : スキャナがエラーを吐いていたらプログラムを修正させる
func scanErrPanic(scanner *bufio.Scanner, errorMessage string) {
	if err := scanner.Err(); err != nil {
		fmt.Printf("[ERROR] %s", errorMessage)
		panic(err)
	}
}

// mallocFmoInfo : ConnectivityInfo&ResidueName以外のFmoInfo構造体のスライスを初期化しておく。
func (fmoInfo *FmoInfo) mallocFmoInfo() {
	if fmoInfo.FragmentNum < 1 || fmoInfo.AtomNum < 1 {
		panic("[ERROR] Number of fragments or number of atoms are illegal.")
	}
	fmoInfo.Atom = make([]atominfo.AtomInfo, fmoInfo.AtomNum)
	fmoInfo.ResidueInFragment = make([]map[int]struct{}, fmoInfo.FragmentNum)

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
		fmoInfo.ResidueInFragment[i] = make(map[int]struct{})
	}
	fmoInfo.ResidueName = make(map[int]string)
}

func (fmoInfo *FmoInfo) createConnectivityMatrix() {
	fmoInfo.ConnectivityMatrix = make([][]bool, fmoInfo.FragmentNum)
	for i := 0; i < fmoInfo.FragmentNum; i++ {
		fmoInfo.ConnectivityMatrix[i] = make([]bool, fmoInfo.FragmentNum)
	}
	for i := 0; i < fmoInfo.ConnectivityNum; i++ {
		// 原子の結合情報をフラグメントごとの結合情報に変換
		src := fmoInfo.Atom[fmoInfo.ConnectivityInfo[i].dest-1].FragmentNum
		dest := fmoInfo.Atom[fmoInfo.ConnectivityInfo[i].src-1].FragmentNum
		fmoInfo.ConnectivityMatrix[src-1][dest-1] = true
		fmoInfo.ConnectivityMatrix[dest-1][src-1] = true
	}
}

// HartreeToKcalPerMol : 単位をHartreeからKcal/molに変換
func HartreeToKcalPerMol(val float64) float64 {
	return val * hartreeKcalPerMol
}

// SetUnitHartree : FmoInfoが出力するときの単位をHartreeに変更する。
func (fmoInfo *FmoInfo) SetUnitHartree() {
	fmoInfo.unit = Hartree
}

// SetUnitKcalPerMol : FmoInfoが出力するときの単位をKcalPerMolに変更する。
func (fmoInfo *FmoInfo) SetUnitKcalPerMol() {
	fmoInfo.unit = KcalPerMol
}

// GenerateCSV : IFIEをCSVに出力する
func (fmoInfo *FmoInfo) GenerateCSV(path string) {
	fp, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	// csvに出力する文字列を格納する2次元スライス
	csvdata := make([][]string, fmoInfo.FragmentNum+1)
	for i := 0; i < fmoInfo.FragmentNum+1; i++ {
		csvdata[i] = make([]string, fmoInfo.FragmentNum+1)
	}
	writer := csv.NewWriter(fp)

	// csv1行目
	i := 1
	for _, fragment := range fmoInfo.ResidueInFragment {
		for residueNum := range fragment {
			csvdata[0][i] += fmoInfo.ResidueName[residueNum] + strconv.Itoa(residueNum) + " "
		}
		i++
	}

	// csv2行目以降
	i = 0
	writer.Write(csvdata[0]) // csv1行目出力
	for _, fragment := range fmoInfo.ResidueInFragment {
		for residueNum := range fragment {
			csvdata[i+1][0] += fmoInfo.ResidueName[residueNum] + strconv.Itoa(residueNum) + " "
		}
		for j := 0; j < fmoInfo.FragmentNum; j++ {
			switch fmoInfo.unit {
			case Hartree:
				csvdata[i+1][j+1] = strconv.FormatFloat(fmoInfo.Ifie[i][j], 'f', 15, 64)
			case KcalPerMol:
				csvdata[i+1][j+1] = strconv.FormatFloat(HartreeToKcalPerMol(fmoInfo.Ifie[i][j]), 'f', 15, 64)
			}
		}
		i++
		writer.Write(csvdata[i])
		writer.Flush()
	}
	fmt.Println("Wrote Successful.")
}

// LoadCPF : CPFからデータを読み込んでFmoInfo型に格納
func (fmoInfo *FmoInfo) LoadCPF(path string) {
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

	fmoInfo.createConnectivityMatrix()

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

			// 結合しているフラグメント間のIFIEは共有結合エネルギーになってしまうそうなので0にする
			if !fmoInfo.ConnectivityMatrix[i][j] {
				fmoInfo.Ifie[i][j] = fmoInfo.NuclearRepulsionEnergy[i][j] + fmoInfo.HfElectronEnergy[i][j] + fmoInfo.Mp2Ifie[i][j]
				fmoInfo.Ifie[j][i] = fmoInfo.Ifie[i][j]
			}
		}
	}

	/*
		for i := 0; i < fmoInfo.FragmentNum; i++ {
			for j := 0; j < fmoInfo.FragmentNum; j++ {
				fmt.Printf("%20.14f", fmoInfo.Ifie[i][j]*627.5095)
				//fmt.Printf("%t ", fmoInfo.ConnectivityMatrix[i][j])
			}
			fmt.Println()
		}
	*/

	fmt.Println("Loaded successful.")
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

	fmoInfo.mallocFmoInfo()

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
		//fmoInfo.Atom[i].PrintAtomInfo()
		fmoInfo.ResidueInFragment[fmoInfo.Atom[i].FragmentNum-1][fmoInfo.Atom[i].ResidueNum] = struct{}{}
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
	}
}
