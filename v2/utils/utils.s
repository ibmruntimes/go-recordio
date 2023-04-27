#include "go_asm.h"
#include "textflag.h"

TEXT ·IefssreqX(SB), NOSPLIT|NOFRAME, $0
	MOVD g, R8                // R13-> R8
	MOVD R14, R9              // R14-> R9
	MOVD parm+0(FP), R1       // parm-> R1
	MOVD branch_ptr+8(FP), R7 // branch_ptr->R7
	MOVD dsa+16(FP), g        // dsa-> R13
	MOVD R15, R10             // R15-> R10
	MOVD R7, R15              // branch_ptr -> R15
	BYTE $0x01; BYTE $0x0d    // SAM31
	BYTE $0x05; BYTE $0xef    // BALR 14,15 branch to IEFSSREQ
	BYTE $0x01; BYTE $0x0e    // SAM64
	MOVD R15, R7              // R15-> R7  (return value)
	MOVD R10, R15             // restore R15 (so that FP is valid)
	MOVD R7, ret+24(FP)       // set return value
	MOVD R8, g                // restore R13
	MOVD R9, R14              // restore R14
	RET

TEXT ·Bpxcall(SB), NOSPLIT|NOFRAME, $0
	MOVD  plist_base+0(FP), R1  // r1 points to plist
	MOVD  bpx_offset+24(FP), R2 // r2 offset to BPX vector table
	MOVD  R14, R7               // save r14
	MOVD  R15, R8               // save r15
	MOVWZ 16(R0), R9
	MOVWZ 544(R9), R9
	MOVWZ 24(R9), R9            // call vector in r9
	ADD   R2, R9                // add offset to vector table
	MOVWZ (R9), R9              // r9 points to entry point
	BYTE  $0x0D                 // BL R14,R9 --> basr r14,r9
	BYTE  $0xE9                 // clobbers 0,1,14,15
	MOVD  R8, R15               // restore 15
	JMP   R7                    // return via saved return address

TEXT ·Svc8(SB), NOSPLIT|NOFRAME, $0
	MOVD r0+0(FP), R0      // arg1-> R0
	MOVD r1+8(FP), R1      // arg2-> R1
	MOVD R15, R2           // save r15
	BYTE $0x0A             // SVC 8
	BYTE $0x08             // ...
	MOVD R15, R3           // R15->R3
	MOVD R2, R15           // restore R15
	MOVD R0, retr0+16(FP)
	MOVD R1, retr1+24(FP)
	MOVD R3, retr15+32(FP)
	RET

TEXT ·Svc9(SB), NOSPLIT|NOFRAME, $0
	MOVD r0+0(FP), R0     // arg1-> R0
	MOVD R15, R2          // save r15
	BYTE $0x0A            // SVC 9
	BYTE $0x09            // ...
	MOVD R15, R3          // R15->R3
	MOVD R2, R15          // restore R15
	MOVD R3, retr15+8(FP)
	RET

TEXT ·Call24(SB), NOSPLIT|NOFRAME, $0
	MOVD g, R8                                                             // preserve R13,R14,R15
	MOVD R14, R9
	MOVD R15, R10
	MOVD modinfo+0(FP), R7                                                 // arg1-> R0
	MOVD 16(R7), R1
	MOVD 24(R7), g
	MOVD 8(R7), R15
	BYTE $0x0D                                                             // BASR 14,0
	BYTE $0xE0
	ADD  $22, R14                                                          // suppose to be address of label BACK
	MOVD R14, 64(R7)                                                       // set the branch back adddress
	MOVD $48(R7), R14                                                      // R14 points to SAM24
	BYTE $0xEB; BYTE $0xEC; BYTE $0xD0; BYTE $0x48; BYTE $0x00; BYTE $0x26 // STMH     r14,r12,72(r13) save higher half of register
	BYTE $0x07; BYTE $0xFE                                                 // BR 14

BACK:
	BYTE $0xEB; BYTE $0xEC; BYTE $0xD0; BYTE $0x48; BYTE $0x00; BYTE $0x96 // LMH      r14,r12,72(r13) restore higher half of register
	MOVD R15, 32(R7)                                                       // set p.R15
	MOVD R15, 16(R10)
	MOVD R10, R15
	MOVD R9, R14
	MOVD R8, g
	RET

TEXT ·Call31(SB), NOSPLIT|NOFRAME, $0
	MOVD g, R8                                                             // preserve R13,R14,R15
	MOVD R14, R9
	MOVD R15, R10
	MOVD modinfo+0(FP), R7                                                 // arg1-> R0
	MOVD 16(R7), R1
	MOVD 24(R7), g
	MOVD 8(R7), R15
	BYTE $0xEB; BYTE $0xEC; BYTE $0xD0; BYTE $0x48; BYTE $0x00; BYTE $0x26 // STMH     r14,r12,72(r13) save higher half of register
	BYTE $0x01; BYTE $0x0D                                                 // SAM31
	BYTE $0x0D; BYTE $0xEF                                                 // BASR 14,15
	BYTE $0x01; BYTE $0x0E                                                 // SAM64
	BYTE $0xEB; BYTE $0xEC; BYTE $0xD0; BYTE $0x48; BYTE $0x00; BYTE $0x96 // LMH      r14,r12,72(r13) restore higher half of register
	MOVD R15, 32(R7)                                                       // set p.R15
	MOVD R15, 16(R10)
	MOVD R10, R15
	MOVD R9, R14
	MOVD R8, g
	RET

TEXT ·Call64(SB), NOSPLIT|NOFRAME, $0
	MOVD g, R8                                                             // preserve R13,R14,R15
	MOVD R14, R9
	MOVD R15, R10
	MOVD modinfo+0(FP), R7                                                 // arg1-> R0
	MOVD 16(R7), R1
	MOVD 24(R7), g
	MOVD 8(R7), R15
	BYTE $0x0D; BYTE $0xEF                                                 // BASR 14,15
	MOVD R15, 32(R7)                                                       // set p.R15
	MOVD R15, 16(R10)
	MOVD R10, R15
	MOVD R9, R14
	MOVD R8, g
	RET
