package atominfo

import (
	"fmt"
)

// AtomInfo : 原子情報を入れておく
type AtomInfo struct {
	AtomNum     int // 原子の番号
	AtomName    string
	AtomType    string
	ResidueNum  int
	FragmentNum int
}

// PrintAtomInfo : 原子の情報を表示
func (atom *AtomInfo) PrintAtomInfo() {
	fmt.Printf("%5d %3s %3s %4d %4d\n", atom.AtomNum, atom.AtomName, atom.AtomType, atom.ResidueNum, atom.FragmentNum)
}

func Test() int {
	return 0
}
