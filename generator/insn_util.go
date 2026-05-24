package generator

import (
	"fmt"
	"hash/fnv"
	"log"
	"regexp"
	"strconv"
)

type RM int

var noRMs = []RM{0}
var allVXRMs = []RM{0, 1, 2, 3}
var vxrmNames = map[RM]string{
	allVXRMs[0]: "rnu (round-to-nearest-up)",
	allVXRMs[1]: "rne (round-to-nearest-even)",
	allVXRMs[2]: "rdn (round-down (truncate))",
	allVXRMs[3]: "rod round-to-odd (OR bits into LSB, aka \"jam\")",
}

func (r RM) VXRMString() string {
	return vxrmNames[r]
}

var allFRMs = []RM{0, 1, 2, 3, 4}
var frmNames = map[RM]string{
	allFRMs[0]: "RNE (Round to Nearest, ties to Even)",
	allFRMs[1]: "RTZ (Round towards Zero)",
	allFRMs[2]: "RDN (Round Down, towards -Inf)",
	allFRMs[3]: "RUP (Round Up, towards +Inf)",
	allFRMs[4]: "RMM (Round to Nearest, ties to Max Magnitude)",
}

func (r RM) FRMString() string {
	return frmNames[r]
}

type VXSAT bool

type SEW int

var allSEWs = []SEW{8, 16, 32, 64}

func (i *Insn) floatSEWs() []SEW {
	if i.Option.Float16 {
		return []SEW{16, 32, 64}
	} else {
		return []SEW{32, 64}
	}
}

var validSEWs = map[SEW]struct{}{
	allSEWs[0]: {},
	allSEWs[1]: {},
	allSEWs[2]: {},
	allSEWs[3]: {},
}

func (s SEW) String() string {
	if _, ok := validSEWs[s]; !ok {
		log.Fatalln("unreachable")
	}

	return fmt.Sprintf("e%d", s)
}

type LMUL float32

var allLMULs = []LMUL{LMUL(1) / 8, LMUL(1) / 4, LMUL(1) / 2, 1, 2, 4, 8}
var wideningMULs = []LMUL{LMUL(1) / 8, LMUL(1) / 4, LMUL(1) / 2, 1, 2, 4}
var validLMULs = map[LMUL]struct{}{
	allLMULs[0]: {},
	allLMULs[1]: {},
	allLMULs[2]: {},
	allLMULs[3]: {},
	allLMULs[4]: {},
	allLMULs[5]: {},
	allLMULs[6]: {},
}

func nfieldsLMULs(nfields int) []LMUL {
	var lmuls []LMUL
	for _, lmul := range allLMULs {
		if lmul*LMUL(nfields) > LMUL(8) {
			continue
		}
		lmuls = append(lmuls, lmul)
	}
	return lmuls
}

func (l LMUL) String() string {
	if _, ok := validLMULs[l]; !ok {
		log.Fatalln("unreachable")
	}

	if l < 1 {
		return fmt.Sprintf("mf%d", int(1/l))
	}
	return fmt.Sprintf("m%d", int(l))
}

type VLEN int

func (v VLEN) Valid() bool {
	return 32 <= v && v <= 65536 && v&(v-1) == 0
}

type XLEN int

func (x XLEN) Valid(v VLEN) bool {
	if x == 64 && v < 64 {
		return false
	}
	return x == 32 || x == 64
}

func v0t(mask bool) string {
	if mask {
		return ", v0.t"
	}
	return ""
}

func getEEW(name string) SEW {
	s := regexp.MustCompile(`v.+?(\d+)f*\.v`)
	eew, err := strconv.Atoi(s.FindStringSubmatch(name)[1])
	if err != nil {
		log.Fatalln("unreachable")
	}
	return SEW(eew)
}

func getNfields(name string) int {
	s := regexp.MustCompile(`v.+?seg(\d)e.+?\.v`)
	subs := s.FindStringSubmatch(name)
	if len(subs) < 2 {
		return 1
	}
	nfields, err := strconv.Atoi(subs[1])
	if err != nil {
		return 1
	}
	return nfields
}

func iff[T any](condition bool, t T, f T) T {
	if condition {
		return t
	}
	return f
}

func ta(mask bool) string {
	if mask {
		return "ta"
	}
	return "tu"
}

func ma(mask bool) string {
	if mask {
		return "ma"
	}
	return "mu"
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func getVRegs(lmul1 LMUL, v0 bool, seed string) (int, int, int) {
	if lmul1 < LMUL(1) {
		log.Fatalln("unreachable")
	}

	availableOptions := make([]int, 0)
	for i := iff(v0, 0, int(lmul1)); i < 32; i += int(lmul1) {
		availableOptions = append(availableOptions, i)
	}

	shuffleSlice(availableOptions, int64(len(availableOptions))+int64(hash(seed)))
	return availableOptions[0], availableOptions[1], availableOptions[2]
}
