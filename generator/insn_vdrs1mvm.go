package generator

import (
	"fmt"
	"strings"
	"math"
)

func (i *Insn) genCodeVdRs1mVm(pos int) []string {
	nfields := getNfields(i.Name)
	combinations := i.combinations(
		nfieldsLMULs(nfields),
		[]SEW{getEEW(i.Name)},
		[]bool{false, true},
		i.rms(),
	)
	res := make([]string, 0, len(combinations))

	for _, c := range combinations[pos:] {
		builder := strings.Builder{}
		var sew SEW = 8 // sew is data width, c.SEW is offset width
		for ; sew <= SEW(i.Option.XLEN); sew <<= 1 {
			if int(sew) > int(float64(i.Option.XLEN)*float64(c.LMUL)) {
				continue
			}
			emul := (float64(c.SEW) / float64(sew)) * float64(c.LMUL) * float64(nfields)
			if emul > 8 || emul < 1./8 {
				continue
			}
			emul = math.Max(emul, 1) // offset lmul
			builder.WriteString(c.initialize())

			builder.WriteString(i.gWriteRandomData(LMUL(1)))
			builder.WriteString(i.gLoadDataIntoRegisterGroup(0, LMUL(1), SEW(32)))

			lmul1 := LMUL(math.Max((float64(c.SEW)/float64(sew))*float64(c.LMUL), 1))
			vd := int(lmul1)
			
			builder.WriteString(i.gWriteIntegerTestData(lmul1*LMUL(nfields), sew, 0))			
			for nf := 0; nf < nfields; nf++ {
				builder.WriteString(i.gLoadDataIntoRegisterGroup(vd+(nf*int(lmul1)), lmul1, sew))
				builder.WriteString(fmt.Sprintf("li a5, %d\n", i.vlenb()*int(lmul1)))
				builder.WriteString("add a0, a0, a5\n")
			}

			builder.WriteString("# -------------- TEST BEGIN --------------\n")
			builder.WriteString(i.gVsetvli(c.Vl, sew, c.LMUL))
			builder.WriteString(fmt.Sprintf("%s v%d, (a0)%s\n", i.Name, vd, v0t(c.Mask)))
			builder.WriteString("# -------------- TEST END   --------------\n")

			builder.WriteString(i.gResultDataAddr())
			for nf := 0; nf < nfields; nf++ {
				builder.WriteString(i.gStoreRegisterGroupIntoResultData(vd+(nf*int(lmul1)), lmul1, sew))
				builder.WriteString(i.gMagicInsn(vd+(nf*int(lmul1)), lmul1))
			}
		}

		res = append(res, builder.String())
	}

	return res
}
