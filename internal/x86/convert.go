package x86

import (
	"fmt"

	"github.com/tsavola/wag/internal/gen"
	"github.com/tsavola/wag/internal/links"
	"github.com/tsavola/wag/internal/opers"
	"github.com/tsavola/wag/internal/regs"
	"github.com/tsavola/wag/internal/types"
	"github.com/tsavola/wag/internal/values"
)

func (mach X86) ConversionOp(code gen.RegCoder, oper uint16, resultType types.T, source values.Operand) (result values.Operand) {
	if oper == opers.Wrap {
		source.Type = types.I32 // short mov; useful zeroExt flag
		reg, zeroExt := mach.opMaybeResultReg(code, source, false)
		return values.TempRegOperand(resultType, reg, zeroExt)
	}

	reg, zeroExt := mach.opMaybeResultReg(code, source, false)

	switch oper {
	case opers.ExtendS:
		Movsxd.opFromReg(code, 0, reg, reg)
		result = values.TempRegOperand(resultType, reg, false)

	case opers.ExtendU:
		if !zeroExt {
			Mov.opFromReg(code, types.I32, reg, reg)
		}
		result = values.TempRegOperand(resultType, reg, false)

	case opers.Mote:
		Cvts2sSSE.opFromReg(code, source.Type, reg, reg)
		result = values.TempRegOperand(resultType, reg, false)

	case opers.TruncS:
		CvttsSSE2si.opReg(code, source.Type, resultType, regResult, reg)
		code.FreeReg(source.Type, reg)
		result = values.TempRegOperand(resultType, regResult, false)

	case opers.TruncU:
		if resultType == types.I32 {
			CvttsSSE2si.opReg(code, source.Type, types.I64, regResult, reg) // larger target size
		} else {
			panic(fmt.Errorf("%s.trunc_u/%s not implemented", resultType, source.Type))
		}
		code.FreeReg(source.Type, reg)
		result = values.TempRegOperand(resultType, regResult, false)

	case opers.ConvertS:
		Cvtsi2sSSE.opReg(code, resultType, source.Type, regResult, reg)
		code.FreeReg(source.Type, reg)
		result = values.TempRegOperand(resultType, regResult, false)

	case opers.ConvertU:
		if source.Type == types.I32 {
			if !zeroExt {
				Mov.opFromReg(code, types.I32, reg, reg)
			}
			Cvtsi2sSSE.opReg(code, resultType, types.I64, regResult, reg)
		} else {
			mach.opConvertUnsignedI64ToFloat(code, resultType, reg)
		}
		code.FreeReg(source.Type, reg)
		result = values.TempRegOperand(resultType, regResult, false)

	case opers.Reinterpret:
		if source.Type.Category() == types.Int {
			MovSSE.opFromReg(code, source.Type, regResult, reg)
		} else {
			MovSSE.opToReg(code, source.Type, regResult, reg)
		}
		code.FreeReg(source.Type, reg)
		result = values.TempRegOperand(resultType, regResult, true)
	}

	return
}

func (mach X86) opConvertUnsignedI64ToFloat(code gen.Coder, floatType types.T, intReg regs.R) {
	var intReg2 regs.R

	if intReg == regScratch {
		intReg2 = regResult
	} else {
		intReg2 = regScratch
	}

	// this algorithm is copied from code generated by gcc and clang:

	var done links.L
	var huge links.L

	Test.opFromReg(code, types.I64, intReg, intReg)
	Js.rel8.opStub(code)
	huge.AddSite(code.Len())

	// max. 63-bit value
	Cvtsi2sSSE.opReg(code, floatType, types.I64, regResult, intReg)

	JmpRel.rel8.opStub(code)
	done.AddSite(code.Len())

	huge.Addr = code.Len()
	mach.updateBranches8(code, &huge)

	// 64-bit value
	Mov.opFromReg(code, types.I64, intReg2, intReg)
	And.opImm(code, types.I64, intReg2, 1)
	ShrImm.op(code, types.I64, intReg, 1)
	Or.opFromReg(code, types.I64, intReg, intReg2)
	Cvtsi2sSSE.opReg(code, floatType, types.I64, regResult, intReg)
	AddsSSE.opFromReg(code, floatType, regResult, regResult)

	done.Addr = code.Len()
	mach.updateBranches8(code, &done)
}