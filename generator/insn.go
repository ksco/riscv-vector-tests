package generator

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

type insnFormat string

type Option struct {
	VLEN     VLEN
	XLEN     XLEN
	HasV	 bool
	Fp       bool
	Repeat   int
	Float16  bool
	Bfloat16 bool
}

const minStride = -1 // Must be negative
const maxStride = 3  // Must be greater than 1
const strides = maxStride - minStride

type TestData struct {
	CurrentOffset uint64
	VLEN          VLEN

	Raws [][]byte
}

func (t *TestData) Append(raw []byte) uint64 {
	offset := t.CurrentOffset
	t.Raws = append(t.Raws, raw)
	t.CurrentOffset += uint64(len(raw))
	return offset
}

func (t *TestData) String() string {
	builder := strings.Builder{}
	dir := ".quad"
	size := 8
	if t.VLEN == 32 {
		dir = ".word"
		size = 4
	}
	for _, raw := range t.Raws {
		for len(raw) >= size {
			var val uint64
			if size == 4 {
				val = uint64(binary.LittleEndian.Uint32(raw[:4]))
			} else {
				val = binary.LittleEndian.Uint64(raw[:8])
			}
			fmt.Fprintf(&builder, "  %s 0x%x\n", dir, val)
			raw = raw[size:]
		}
	}
	return builder.String()
}

type Insn struct {
	Name         string     `toml:"name"`
	Format       insnFormat `toml:"format"`
	NoTestfloat3 bool       `toml:"notestfloat3"`
	Vxrm         bool       `toml:"vxrm"`
	Vxsat        bool       `toml:"vxsat"`
	Tests        tests      `toml:"tests"`
	Option       Option     `toml:"-"`
	EGW          int        `toml:"-"`
	TestData     *TestData
}

const (
	insnFormatVdRs1mVm     insnFormat = "vd,(rs1),vm"
	insnFormatVs3Rs1mVm    insnFormat = "vs3,(rs1),vm"
	insnFormatVdRs1m       insnFormat = "vd,(rs1)"
	insnFormatVs3Rs1m      insnFormat = "vs3,(rs1)"
	insnFormatVdRs1mRs2Vm  insnFormat = "vd,(rs1),rs2,vm"
	insnFormatVs3Rs1mRs2Vm insnFormat = "vs3,(rs1),rs2,vm"
	insnFormatVdRs1mVs2Vm  insnFormat = "vd,(rs1),vs2,vm"
	insnFormatVs3Rs1mVs2Vm insnFormat = "vs3,(rs1),vs2,vm"
	insnFormatVdVs2Vs1     insnFormat = "vd,vs2,vs1"
	insnFormatVdVs2Vs1V0   insnFormat = "vd,vs2,vs1,v0"
	insnFormatVdVs2Vs1Vm   insnFormat = "vd,vs2,vs1,vm"
	insnFormatVdVs2Rs1     insnFormat = "vd,vs2,rs1"
	insnFormatVdVs2Rs1V0   insnFormat = "vd,vs2,rs1,v0"
	insnFormatVdVs2Fs1V0   insnFormat = "vd,vs2,fs1,v0"
	insnFormatVdVs2Rs1Vm   insnFormat = "vd,vs2,rs1,vm"
	insnFormatVdVs2Fs1Vm   insnFormat = "vd,vs2,fs1,vm"
	insnFormatVdVs2Imm     insnFormat = "vd,vs2,imm"
	insnFormatVdVs2ImmV0   insnFormat = "vd,vs2,imm,v0"
	insnFormatVdVs2ImmVm   insnFormat = "vd,vs2,imm,vm"
	insnFormatVdVs2Uimm    insnFormat = "vd,vs2,uimm"
	insnFormatVdVs2UimmVm  insnFormat = "vd,vs2,uimm,vm"
	insnFormatVdVs1Vs2Vm   insnFormat = "vd,vs1,vs2,vm"
	insnFormatVdRs1Vs2Vm   insnFormat = "vd,rs1,vs2,vm"
	insnFormatVdFs1Vs2Vm   insnFormat = "vd,fs1,vs2,vm"
	insnFormatVdVs1        insnFormat = "vd,vs1"
	insnFormatVdRs1        insnFormat = "vd,rs1"
	insnFormatVdFs1        insnFormat = "vd,fs1"
	insnFormatVdImm        insnFormat = "vd,imm"
	insnFormatVdVs2        insnFormat = "vd,vs2"
	insnFormatVdVs2Vm      insnFormat = "vd,vs2,vm"
	insnFormatVdVs2VmP2    insnFormat = "vd,vs2,vm/2"
	insnFormatVdVs2VmP3    insnFormat = "vd,vs2,vm/3"
	insnFormatRdVs2Vm      insnFormat = "rd,vs2,vm"
	insnFormatRdVs2        insnFormat = "rd,vs2"
	insnFormatFdVs2        insnFormat = "fd,vs2"
	insnFormatVdVm         insnFormat = "vd,vm"
	insnFormatVsetvli      insnFormat = "vsetvli"
	insnFormatVsetvl       insnFormat = "vsetvl"
	insnFormatVsetivli     insnFormat = "vsetivli"
)

var formats = map[insnFormat]struct{}{
	insnFormatVdRs1mVm:     {},
	insnFormatVs3Rs1mVm:    {},
	insnFormatVdRs1m:       {},
	insnFormatVs3Rs1m:      {},
	insnFormatVdRs1mRs2Vm:  {},
	insnFormatVs3Rs1mRs2Vm: {},
	insnFormatVdRs1mVs2Vm:  {},
	insnFormatVs3Rs1mVs2Vm: {},
	insnFormatVdVs2Vs1:     {},
	insnFormatVdVs2Vs1V0:   {},
	insnFormatVdVs2Vs1Vm:   {},
	insnFormatVdVs2Rs1:     {},
	insnFormatVdVs2Rs1V0:   {},
	insnFormatVdVs2Fs1V0:   {},
	insnFormatVdVs2Rs1Vm:   {},
	insnFormatVdVs2Fs1Vm:   {},
	insnFormatVdVs2Imm:     {},
	insnFormatVdVs2ImmV0:   {},
	insnFormatVdVs2ImmVm:   {},
	insnFormatVdVs2Uimm:    {},
	insnFormatVdVs2UimmVm:  {},
	insnFormatVdVs1Vs2Vm:   {},
	insnFormatVdRs1Vs2Vm:   {},
	insnFormatVdFs1Vs2Vm:   {},
	insnFormatVdVs1:        {},
	insnFormatVdRs1:        {},
	insnFormatVdFs1:        {},
	insnFormatVdImm:        {},
	insnFormatVdVs2:        {},
	insnFormatVdVs2Vm:      {},
	insnFormatVdVs2VmP2:    {},
	insnFormatVdVs2VmP3:    {},
	insnFormatRdVs2Vm:      {},
	insnFormatRdVs2:        {},
	insnFormatFdVs2:        {},
	insnFormatVdVm:         {},
	insnFormatVsetvli:      {},
	insnFormatVsetvl:       {},
	insnFormatVsetivli:     {},
}

func (i *Insn) genCodeCombinations(pos int) []string {
	switch i.Format {
	case insnFormatVdRs1mVm:
		return i.genCodeVdRs1mVm(pos)
	case insnFormatVs3Rs1mVm:
		return i.genCodeVs3Rs1mVm(pos)
	case insnFormatVdRs1m:
		return i.genCodeVdRs1m(pos)
	case insnFormatVs3Rs1m:
		return i.genCodeVs3Rs1m(pos)
	case insnFormatVdRs1mRs2Vm:
		return i.genCodeVdRs1mRs2Vm(pos)
	case insnFormatVs3Rs1mRs2Vm:
		return i.genCodeVs3Rs1mRs2Vm(pos)
	case insnFormatVdRs1mVs2Vm:
		return i.genCodeVdRs1mVs2Vm(pos)
	case insnFormatVs3Rs1mVs2Vm:
		return i.genCodeVs3Rs1mVs2Vm(pos)
	case insnFormatVdVs2Vs1:
		return i.genCodeVdVs2Vs1(pos)
	case insnFormatVdVs2Vs1V0:
		return i.genCodeVdVs2Vs1V0(pos)
	case insnFormatVdVs2Rs1:
		return i.genCodeVdVs2Rs1(pos)
	case insnFormatVdVs2Rs1V0:
		return i.genCodeVdVs2Rs1V0(pos)
	case insnFormatVdVs2Fs1V0:
		return i.genCodeVdVs2Fs1V0(pos)
	case insnFormatVdVs2Vs1Vm:
		return i.genCodeVdVs2Vs1Vm(pos)
	case insnFormatVdVs2Rs1Vm:
		return i.genCodeVdVs2Rs1Vm(pos)
	case insnFormatVdVs1Vs2Vm:
		return i.genCodeVdVs1Vs2Vm(pos)
	case insnFormatVdRs1Vs2Vm:
		return i.genCodeVdRs1Vs2Vm(pos)
	case insnFormatVdVs2Fs1Vm:
		return i.genCodeVdVs2Fs1Vm(pos)
	case insnFormatVdVs2Imm:
		return i.genCodeVdVs2Imm(pos)
	case insnFormatVdVs2ImmV0:
		return i.genCodeVdVs2ImmV0(pos)
	case insnFormatVdVs2ImmVm:
		return i.genCodeVdVs2ImmVm(pos)
	case insnFormatVdVs2Uimm:
		return i.genCodeVdVs2Uimm(pos)
	case insnFormatVdVs2UimmVm:
		return i.genCodeVdVs2UimmVm(pos)
	case insnFormatVdFs1Vs2Vm:
		return i.genCodeVdFs1Vs2Vm(pos)
	case insnFormatVdVs1:
		return i.genCodeVdVs1(pos)
	case insnFormatVdRs1:
		return i.genCodeVdRs1(pos)
	case insnFormatVdFs1:
		return i.genCodeVdFs1(pos)
	case insnFormatVdImm:
		return i.genCodeVdImm(pos)
	case insnFormatVdVs2:
		return i.genCodeVdVs2(pos)
	case insnFormatVdVs2Vm:
		return i.genCodeVdVs2Vm(pos)
	case insnFormatVdVs2VmP2:
		return i.genCodeVdVs2VmP2(pos)
	case insnFormatVdVs2VmP3:
		return i.genCodeVdVs2VmP3(pos)
	case insnFormatRdVs2Vm:
		return i.genCodeRdVs2Vm(pos)
	case insnFormatRdVs2:
		return i.genCodeRdVs2(pos)
	case insnFormatFdVs2:
		return i.genCodeFdVs2(pos)
	case insnFormatVdVm:
		return i.genCodeVdVm(pos)
	case insnFormatVsetvli:
		return i.genCodevsetvli(pos)
	case insnFormatVsetvl:
		return i.genCodevsetvl(pos)
	case insnFormatVsetivli:
		return i.genCodevsetivli(pos)
	default:
		log.Fatalln("unreachable")
		return nil
	}
}

func ReadInsnFromToml(contents []byte, option Option) (*Insn, error) {
	i := Insn{
		Option:   option,
		EGW:      0,
		TestData: &TestData{VLEN: option.VLEN},
	}

	if err := i.check(); err != nil {
		return nil, err
	}

	if err := toml.Unmarshal(contents, &i); err != nil {
		return nil, err
	}

	if err := i.Tests.initialize(); err != nil {
		return nil, err
	}

	if _, ok := formats[i.Format]; !ok {
		return nil, errors.New("invalid test format")
	}

	return &i, nil
}

func (i *Insn) check() error {
	if !i.Option.VLEN.Valid() {
		return fmt.Errorf("wrong VLEN: %d", i.Option.VLEN)
	}

	if !i.Option.XLEN.Valid(i.Option.VLEN) {
		return fmt.Errorf("wrong XLEN: %d", i.Option.XLEN)
	}
	return nil
}

func (i *Insn) Generate(splitPerLines int) []string {
	res := make([]string, 0)

	codes, data := i.genMergedCodeCombinations(splitPerLines)
	for idx, code := range codes {
		builder := strings.Builder{}
		builder.WriteString(i.genHeader())
		builder.WriteString(code)
		builder.WriteString(data[idx])
		res = append(res, builder.String())
	}
	return res
}

func (i *Insn) genHeader() string {
	return fmt.Sprintf(`#
# This file is automatically generated. Do not edit.
# Instruction: %s

#include "riscv_test.h"
#include "test_macros.h"

RVTEST_RV%dUV%s
`, i.Name, i.Option.XLEN, iff(i.Option.Fp, "", "X"))
}

func (i *Insn) genMergedCodeCombinations(splitPerLines int) ([]string, []string) {
	codeRes := make([]string, 0)
	dataRes := make([]string, 0)
	builder := strings.Builder{}
	done := false
	pos := 0
	for !done {
		cs := i.genCodeCombinations(pos)
		if len(cs) == 0 {
			break
		}
		for idx, c := range cs {
			builder.WriteString(c)
			str := builder.String()
			if (splitPerLines > 0 && strings.Count(str, "\n") > splitPerLines) ||
				idx == len(cs)-1 {
				if len(str) != 0 {
					buf := fmt.Sprintf(`
RVTEST_CODE_BEGIN

# Zero all vector registers
vsetvli t0, x0, e8,m8,tu,mu
vmv.v.i v0, 0x0
vmv.v.i v8, 0x0
vmv.v.i v16, 0x0
vmv.v.i v24, 0x0

%s
  TEST_CASE(2, x0, 0x0)
  TEST_PASSFAIL
RVTEST_CODE_END
`, str)
					codeRes = append(codeRes, buf)
					dataRes = append(dataRes, i.genData())
				}
				i.TestData = &TestData{VLEN: i.Option.VLEN}
				builder.Reset()
				pos += idx
				if idx == len(cs)-1 {
					done = true
				}
				break
			}
		}
	}

	return codeRes, dataRes
}

func (i *Insn) genData() string {
	dataSize := i.vlenb() * 8 /* max LMUL */ * 8 /* max NFIELDS */
	// Stride insns
	if strings.HasPrefix(i.Name, "vlse") ||
		strings.HasPrefix(i.Name, "vsse") ||
		strings.HasPrefix(i.Name, "vlsse") ||
		strings.HasPrefix(i.Name, "vssse") {
		dataSize *= strides
	} else if strings.HasPrefix(i.Name, "vw") ||
		strings.HasPrefix(i.Name, "vfw") {
		dataSize *= 2
	}

	return fmt.Sprintf(`
  .data
RVTEST_DATA_BEGIN

# Reserve space for test data.
resultdata:
  .zero %d

testdata:
%s

RVTEST_DATA_END
`, dataSize, i.TestData.String())
}

func (i *Insn) vlenb() int {
	return int(i.Option.VLEN) / 8
}

func (i *Insn) integerTestCases(sew SEW) [][]any {
	return i.testCases(false, sew)
}

func (i *Insn) testCases(float bool, sew SEW) [][]any {
	res := make([][]any, 0)
	if !float {
		for _, c := range i.Tests.Base {
			l := make([]any, len(c))
			for b, op := range c {
				l[b] = op
			}
			res = append(res, l)
		}
	}

	switch sew {
	case 8:
		for _, c := range i.Tests.SEW8 {
			l := make([]any, len(c))
			for b, op := range c {
				l[b] = op
			}
			res = append(res, l)
		}
	case 16:
		if float {
			var testCases []testCase[uint16]

			if i.Option.Bfloat16 {
				testCases = i.Tests.BF16SEW16
			} else {
				testCases = i.Tests.FSEW16
			}

			for _, c := range testCases {
				l := make([]any, len(c))
				for b, op := range c {
					l[b] = op
				}
				res = append(res, l)
			}
			break
		}
		for _, c := range i.Tests.SEW16 {
			l := make([]any, len(c))
			for b, op := range c {
				l[b] = op
			}
			res = append(res, l)
		}
	case 32:
		if float {
			for _, c := range i.Tests.FSEW32 {
				l := make([]any, len(c))
				for b, op := range c {
					l[b] = math.Float32bits(op)
				}
				res = append(res, l)
			}
			break
		}
		for _, c := range i.Tests.SEW32 {
			l := make([]any, len(c))
			for b, op := range c {
				l[b] = op
			}
			res = append(res, l)
		}
	case 64:
		if float {
			for _, c := range i.Tests.FSEW64 {
				l := make([]any, len(c))
				for b, op := range c {
					l[b] = math.Float64bits(op)
				}
				res = append(res, l)
			}
			break
		}
		for _, c := range i.Tests.SEW64 {
			l := make([]any, len(c))
			for b, op := range c {
				l[b] = op
			}
			res = append(res, l)
		}
	}

	return res
}

type combination struct {
	SEW   SEW
	LMUL  LMUL
	LMUL1 LMUL
	Vl    int
	Mask  bool
	RM    RM
	FRM   bool
	VXSAT VXSAT
}

func (c *combination) initialize() string {
	// write comments, set vxrm/frm, clear vxsat if necessary
	str := fmt.Sprintf(`
# Generating tests for VL: %d, LMUL: %s, SEW: %s, Mask: %v

# Initialize rounding mode CSR
csrwi %s, %d # %s

`,
		c.Vl, c.LMUL.String(), c.SEW.String(), c.Mask, iff(c.FRM, "frm", "vxrm"), c.RM, iff(c.FRM, c.RM.FRMString(), c.RM.VXRMString()))
	if c.VXSAT {
		str = fmt.Sprintf(`%s# Clear vxsat CSR
csrci vxsat, 1

`, str)
	}
	return str
}

func (i *Insn) combinations(lmuls []LMUL, sews []SEW, masks []bool, rms []RM) []combination {
	res := make([]combination, 0)
	for _, lmul := range lmuls {
		if i.EGW > 0 && int(float64(lmul)*float64(i.Option.VLEN)) < i.EGW {
			continue
		}
		for _, sew := range sews {
			if int(i.Option.XLEN) < int(sew) {
				continue
			}
			if float64(lmul) < float64(sew)/float64(i.Option.XLEN) {
				continue
			}
			lmul1 := LMUL(math.Max(float64(lmul), 1))
			for _, mask := range masks {
				vlmax1 := int((float64(i.Option.VLEN) / float64(sew)) * float64(lmul1))
				for _, vl := range []int{0, vlmax1/2 - 1, vlmax1 / 2, vlmax1, vlmax1 - 1, vlmax1 + 1, getRandomInt(vlmax1), getRandomInt(vlmax1), getRandomInt(vlmax1)} {
					for _, rm := range rms {
						res = append(res, combination{
							SEW:   sew,
							LMUL:  lmul,
							LMUL1: lmul1,
							Vl:    vl,
							Mask:  mask,
							RM:    rm,
							FRM:   i.Option.Fp,
							VXSAT: VXSAT(i.Vxsat),
						})
					}
				}
			}
		}
	}

	return res
}

func (i *Insn) rms() []RM {
	if i.Vxrm {
		return allVXRMs
	} else if i.Option.Fp {
		return allFRMs
	}
	return noRMs
}

type vsetvlicombinations struct {
	SEW  SEW
	LMUL LMUL
	vta  bool
	vma  bool
}

func (c *vsetvlicombinations) comment() string {
	return fmt.Sprintf(
		"\n\n# Generating tests for vsetvli: LMUL: %s, SEW: %s, vta: %s, vma: %s\n\n",
		c.LMUL.String(),
		c.SEW.String(),
		ta(c.vta),
		ma(c.vma),
	)
}

func (i *Insn) vsetvlicombinations(lmuls []LMUL, sews []SEW, vtas []bool, vmas []bool) []vsetvlicombinations {
	res := make([]vsetvlicombinations, 0)
	for _, lmul := range lmuls {
		for _, sew := range sews {
			if int(i.Option.XLEN) < int(sew) {
				continue
			}
			if float64(lmul) < float64(sew)/float64(i.Option.XLEN) {
				continue
			}
			for _, vta := range vtas {
				for _, vma := range vmas {
					res = append(res, vsetvlicombinations{
						SEW:  sew,
						LMUL: lmul,
						vta:  vta,
						vma:  vma,
					})
				}
			}
		}
	}
	return res
}

type vtype struct {
	lmul float32
	sew  int
	vta  bool
	vma  bool
}
