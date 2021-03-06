// Copyright (c) 2016 Timo Savola. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codegen

import (
	"github.com/tsavola/wag/internal/gen"
	"github.com/tsavola/wag/internal/gen/debug"
	"github.com/tsavola/wag/internal/gen/link"
	"github.com/tsavola/wag/internal/gen/reg"
	"github.com/tsavola/wag/internal/gen/storage"
	"github.com/tsavola/wag/internal/loader"
	"github.com/tsavola/wag/internal/module"
	"github.com/tsavola/wag/wa"
	"github.com/tsavola/wag/wa/opcode"
)

var (
	errCallParamsExceedStack = module.Error("function call parameter count exceeds stack operand count")
	errCallParamTypeMismatch = module.Error("function call argument has wrong type")
)

func genCall(f *gen.Func, load loader.L, op opcode.Opcode, info opInfo) (deadend bool) {
	opSaveOperands(f)

	funcIndex := load.Varuint32()

	if f.ImportContext != nil {
		opCallInImportFunc(f, funcIndex)
	} else {
		opCallInNormalFunc(f, op, funcIndex)
	}
	return
}

func opCallInNormalFunc(f *gen.Func, op opcode.Opcode, funcIndex uint32) {
	if funcIndex >= uint32(len(f.Module.Funcs)) {
		panic(module.Errorf("%s: function index out of bounds: %d", op, funcIndex))
	}

	sig := f.Module.Types[f.Module.Funcs[funcIndex]]
	checkCallOperandCount(f, sig)
	opCall(f, &f.FuncLinks[funcIndex].L)

	// The called function's initial suspension point was certainly executed.
	if len(f.BranchTargets) > 0 {
		getCurrentBlock(f).Suspension = true
	}

	opFinalizeCall(f, sig)
}

func opCallInImportFunc(f *gen.Func, funcIndex uint32) {
	imp := f.ImportContext.ImportFuncs[funcIndex]
	sigIndex := f.ImportContext.Funcs[funcIndex]
	sig := f.ImportContext.Types[sigIndex]
	checkCallOperandCount(f, sig)

	// See the comment in opBranchIf.
	if !getCurrentBlock(f).WeakDead {
		asm.CallImportVector(&f.Prog, imp.VectorIndex, imp.Variadic, len(sig.Params), int(sigIndex))
	}

	f.MapCallAddr(f.Text.Addr)
	opFinalizeCall(f, sig)
}

func genCallIndirect(f *gen.Func, load loader.L, op opcode.Opcode, info opInfo) (deadend bool) {
	sigIndex := load.Varuint32()
	if sigIndex >= uint32(len(f.Module.Types)) {
		panic(module.Errorf("%s: signature index out of bounds: %d", op, sigIndex))
	}

	load.Byte() // reserved

	funcIndex := popOperand(f, wa.I32)

	opSaveOperands(f)

	var funcIndexReg reg.R
	if funcIndex.Storage == storage.Reg {
		funcIndexReg = funcIndex.Reg()
	} else {
		funcIndexReg = reg.Result
		asm.Move(f, funcIndexReg, funcIndex)
	}

	sig := f.Module.Types[sigIndex]
	checkCallOperandCount(f, sig)
	opCallIndirect(f, int32(sigIndex), funcIndexReg)

	// The called function's initial suspension point was certainly executed.
	if len(f.BranchTargets) > 0 {
		getCurrentBlock(f).Suspension = true
	}

	opFinalizeCall(f, sig)
	return
}

func checkCallOperandCount(f *gen.Func, sig wa.FuncType) {
	if debug.Enabled {
		debug.Printf("sig: %s", sig)
	}

	if len(sig.Params) > f.StackDepth-f.FrameBase {
		panic(errCallParamsExceedStack)
	}

	for i, t := range sig.Params {
		if t != f.Operands[len(f.Operands)-len(sig.Params)+i].Type {
			panic(errCallParamTypeMismatch)
		}
	}
}

func opFinalizeCall(f *gen.Func, sig wa.FuncType) {
	f.Regs.CheckNoneAllocated()

	opDropCallOperands(f, len(sig.Params))
	pushResultRegOperand(f, sig.Result)
}

func opCall(f *gen.Func, l *link.L) {
	if l.Addr != 0 {
		asm.Call(&f.Prog, l.Addr)
	} else {
		asm.CallMissing(&f.Prog, f.AtomicCallStubs)
	}
	f.MapCallAddr(f.Text.Addr)
	if l.Addr == 0 {
		l.AddSite(f.Text.Addr)
	}
}

func opCallIndirect(f *gen.Func, sigIndex int32, funcIndexReg reg.R) {
	asm.CallIndirect(f, sigIndex, funcIndexReg)
	f.MapCallAddr(f.Text.Addr)
}
