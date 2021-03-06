// Copyright 2011 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

TEXT ·CompareAndSwapInt32(SB),7,$0
	JMP	·CompareAndSwapUint32(SB)

TEXT ·CompareAndSwapUint32(SB),7,$0
	MOVQ	valptr+0(FP), BP
	MOVL	old+8(FP), AX
	MOVL	new+12(FP), CX
	LOCK
	CMPXCHGL	CX, 0(BP)
	SETEQ	ret+16(FP)
	RET

TEXT ·CompareAndSwapUintptr(SB),7,$0
	JMP	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapInt64(SB),7,$0
	JMP	·CompareAndSwapUint64(SB)

TEXT ·CompareAndSwapUint64(SB),7,$0
	MOVQ	valptr+0(FP), BP
	MOVQ	old+8(FP), AX
	MOVQ	new+16(FP), CX
	LOCK
	CMPXCHGQ	CX, 0(BP)
	SETEQ	ret+24(FP)
	RET

TEXT ·AddInt32(SB),7,$0
	JMP	·AddUint32(SB)

TEXT ·AddUint32(SB),7,$0
	MOVQ	valptr+0(FP), BP
	MOVL	delta+8(FP), AX
	MOVL	AX, CX
	LOCK
	XADDL	AX, 0(BP)
	ADDL	AX, CX
	MOVL	CX, ret+16(FP)
	RET

TEXT ·AddUintptr(SB),7,$0
	JMP	·AddUint64(SB)

TEXT ·AddInt64(SB),7,$0
	JMP	·AddUint64(SB)

TEXT ·AddUint64(SB),7,$0
	MOVQ	valptr+0(FP), BP
	MOVQ	delta+8(FP), AX
	MOVQ	AX, CX
	LOCK
	XADDQ	AX, 0(BP)
	ADDQ	AX, CX
	MOVQ	CX, ret+16(FP)
	RET
