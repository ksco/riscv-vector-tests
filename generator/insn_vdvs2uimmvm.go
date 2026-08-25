package generator

import (
	"fmt"
	"math"
	"strings"
)

func (i *Insn) genCodeVdVs2UimmVm(pos int) []string {
	vdWidening := strings.HasPrefix(i.Name, "vw")
	vs2Widening := strings.HasSuffix(i.Name, ".wi")
	// Either side 2*SEW (vw* widens vd, .wi widens vs2) caps SEW at ELEN/2;
	// otherwise SEW=64 is legal and must be tested.
	sews := iff(vs2Widening || vdWidening, allSEWs[:len(allSEWs)-1], allSEWs)
	vs2Size := iff(vs2Widening, 2, 1)
	vdSize := iff(vdWidening, 2, 1)

	combinations := i.combinations(
		iff(vdWidening || vs2Widening, wideningMULs, allLMULs),
		sews,
		[]bool{false, true},
		i.rms(),
	)
	res := make([]string, 0, len(combinations))

	for _, c := range combinations[pos:] {
		builder := strings.Builder{}
		builder.WriteString(c.initialize())

		builder.WriteString(i.gWriteRandomData(LMUL(1)))
		builder.WriteString(i.gLoadDataIntoRegisterGroup(0, LMUL(1), SEW(32)))

		vdEMUL1 := LMUL(math.Max(float64(int(c.LMUL)*vdSize), 1))
		vs2EMUL1 := LMUL(math.Max(float64(int(c.LMUL)*vs2Size), 1))
		vdEEW := c.SEW * SEW(vdSize)
		vs2EEW := c.SEW * SEW(vs2Size)
		if vdEEW > SEW(i.Option.XLEN) || vs2EEW > SEW(i.Option.XLEN) {
			res = append(res, "")
			continue
		}

		vd := int(vdEMUL1)
		vs2 := vd * 2

		builder.WriteString(i.gWriteRandomData(c.LMUL1))
		builder.WriteString(i.gLoadDataIntoRegisterGroup(vd, c.LMUL1, SEW(8)))

		builder.WriteString(i.gWriteIntegerTestData(vs2EMUL1, vs2EEW, 1))
		builder.WriteString(i.gLoadDataIntoRegisterGroup(vs2, vs2EMUL1, vs2EEW))

		cases := i.integerTestCases(c.SEW)
		for a := 0; a < len(cases); a++ {
			builder.WriteString("# -------------- TEST BEGIN --------------\n")
			builder.WriteString(i.gVsetvli(c.Vl, c.SEW, c.LMUL))
			switch c.SEW {
			case 8:
				builder.WriteString(fmt.Sprintf("%s v%d, v%d, %d%s\n",
					i.Name, vd, vs2, convNum[uint8](cases[a][0]), v0t(c.Mask)))
			case 16:
				builder.WriteString(fmt.Sprintf("%s v%d, v%d, %d%s\n",
					i.Name, vd, vs2, convNum[uint16](cases[a][0]), v0t(c.Mask)))
			case 32:
				builder.WriteString(fmt.Sprintf("%s v%d, v%d, %d%s\n",
					i.Name, vd, vs2, convNum[uint32](cases[a][0]), v0t(c.Mask)))
			case 64:
				builder.WriteString(fmt.Sprintf("%s v%d, v%d, %d%s\n",
					i.Name, vd, vs2, convNum[uint64](cases[a][0]), v0t(c.Mask)))
			}
			builder.WriteString("# -------------- TEST END   --------------\n")

			builder.WriteString(i.gResultDataAddr())
			builder.WriteString(i.gStoreRegisterGroupIntoResultData(vd, c.LMUL1, c.SEW))
			builder.WriteString(i.gMagicInsn(vd, c.LMUL1))
		}

		res = append(res, builder.String())
	}

	return res
}
