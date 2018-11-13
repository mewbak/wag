// Copyright (c) 2016 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package x86

import (
	"github.com/tsavola/wag/internal/gen"
	"github.com/tsavola/wag/internal/gen/condition"
	"github.com/tsavola/wag/internal/gen/link"
	"github.com/tsavola/wag/internal/gen/operand"
	"github.com/tsavola/wag/internal/gen/reg"
	"github.com/tsavola/wag/internal/gen/rodata"
	"github.com/tsavola/wag/internal/isa/prop"
	"github.com/tsavola/wag/internal/isa/x86/in"
	"github.com/tsavola/wag/wa"
)

func (MacroAssembler) Unary(f *gen.Func, props uint16, x operand.O) operand.O {
	switch uint8(props) {
	case prop.IntEqz:
		r, _ := getScratchReg(f, x)
		in.TEST.RegReg(&f.Text, x.Type, r, r)
		f.Regs.Free(x.Type, r)
		return operand.Flags(condition.Eq)

	case prop.IntClz:
		r, _ := allocResultReg(f, x)
		if haveLZCNT() {
			in.LZCNT.RegReg(&f.Text, x.Type, r, r)
		} else {
			in.BSR.RegReg(&f.Text, x.Type, RegScratch, r)
			in.MOVi.RegImm32(&f.Text, x.Type, r, -1)
			in.CMOVE.RegReg(&f.Text, x.Type, RegScratch, r)
			in.MOVi.RegImm32(&f.Text, x.Type, r, (int32(x.Type.Size())<<3)-1)
			in.SUB.RegReg(&f.Text, x.Type, r, RegScratch)
		}
		return operand.Reg(x.Type, r)

	case prop.IntCtz:
		r, _ := allocResultReg(f, x)
		if haveTZCNT() {
			in.TZCNT.RegReg(&f.Text, x.Type, r, r)
		} else {
			in.BSF.RegReg(&f.Text, x.Type, r, r)
			in.MOVi.RegImm32(&f.Text, x.Type, RegScratch, int32(x.Type.Size())<<3)
			in.CMOVE.RegReg(&f.Text, x.Type, r, RegScratch)
		}
		return operand.Reg(x.Type, r)

	case prop.IntPopcnt:
		var r reg.R
		if havePOPCNT() {
			r, _ = allocResultReg(f, x)
			in.POPCNT.RegReg(&f.Text, x.Type, r, r)
		} else {
			r = popcnt(f, x)
		}
		return operand.Reg(x.Type, r)

	case prop.FloatAbs:
		r, _ := allocResultReg(f, x)
		absFloatReg(&f.Prog, x.Type, r)
		return operand.Reg(x.Type, r)

	case prop.FloatNeg:
		r, _ := allocResultReg(f, x)
		negFloatReg(&f.Prog, x.Type, r)
		return operand.Reg(x.Type, r)

	case prop.FloatRoundOp:
		r, _ := allocResultReg(f, x)
		roundMode := int8(props >> 8)
		in.ROUNDSSD.RegRegImm8(&f.Text, x.Type, r, r, roundMode)
		return operand.Reg(x.Type, r)

	default: // FloatSqrt
		r, _ := allocResultReg(f, x)
		in.SQRTSSD.RegReg(&f.Text, x.Type, r, r)
		return operand.Reg(x.Type, r)
	}
}

// absFloatReg in-place.
func absFloatReg(p *gen.Prog, t wa.Type, r reg.R) {
	absMaskAddr := rodata.MaskAddr(rodata.Mask7fBase, t)
	in.ANDPSD.RegMemDisp(&p.Text, t, r, in.BaseText, absMaskAddr)
}

// negFloatReg in-place.
func negFloatReg(p *gen.Prog, t wa.Type, r reg.R) {
	signMaskAddr := rodata.MaskAddr(rodata.Mask80Base, t)
	in.XORPSD.RegMemDisp(&p.Text, t, r, in.BaseText, signMaskAddr)
}

// Population count algorithm:
//
//     func popcnt(x uint) (count int) {
//         for count = 0; x != 0; count++ {
//             x &= x - 1
//         }
//         return
//     }
//
func popcnt(f *gen.Func, x operand.O) (count reg.R) {
	var skip link.L

	count = f.Regs.AllocResult(x.Type)
	pop, _ := getScratchReg(f, x)
	temp := RegZero // cleared again at the end

	in.XOR.RegReg(&f.Text, wa.I32, count, count)

	in.TEST.RegReg(&f.Text, x.Type, pop, pop)
	in.JEcb.Stub8(&f.Text)
	skip.AddSite(f.Text.Addr)

	loopAddr := f.Text.Addr
	in.INC.Reg(&f.Text, wa.I32, count)
	in.MOV.RegReg(&f.Text, x.Type, temp, pop)
	in.DEC.Reg(&f.Text, x.Type, temp)
	in.AND.RegReg(&f.Text, x.Type, pop, temp)
	in.JNEcb.Addr8(&f.Text, loopAddr)

	skip.Addr = f.Text.Addr
	isa.UpdateNearBranches(f.Text.Bytes(), &skip)

	in.XOR.RegReg(&f.Text, wa.I32, RegZero, RegZero) // temp reg
	f.Regs.Free(x.Type, pop)
	return
}
