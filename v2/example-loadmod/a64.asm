PGM CSECT
PGM AMODE 64
PGM RMODE 31
    SYSSTATE AMODE64=YES
    STMG  14,12,SAVF4SAG64RS14-SAVF4SA(13) 
    CNOP  0,4
    BRAS  12,*+8
    DC    A(STATIC_DATA)
    L     12,0(12,0)   
    USING STATIC_DATA,12
    LGR   2,1
    GETMAIN RU,LV=144  
    STG   13,SAVF4SAPREV-SAVF4SA(,1) 
    STG   1,SAVF4SANEXT-SAVF4SA(,13) 
    MVC   SAVF4SAID-SAVF4SA(4,1),=A(SAVF4SAID_VALUE)
    LGR   13,1    
* begin
    GETMAIN RU,LV=wto_size
    USING wto_b,1
    LG    4,0(2) first arg, ptr
    LG    5,8(2) second arg
    LGF   5,0(5) length
    CGHI  5,252
    JNH   OK
    LGHI  5,252
OK  DS    0H
    LGR   6,5
    AGHI  6,4
    STH   6,wto_len
    LA    6,wto_msg
    LGR   7,5
    MVCL  6,4
    XGR   5,5
    STH   5,wto_flag
    LGR   2,1
    SVC   35
    FREEMAIN RU,A=(2),LV=wto_size
* end
    LGR   1,13          
    LG    13,SAVF4SAPREV-SAVF4SA(,13)  
    FREEMAIN RU,A=(1),LV=144 
    SLR   15,15 
    LG    14,SAVF4SAG64RS14-SAVF4SA(,13) 
    LMG   2,12,SAVF4SAG64RS2-SAVF4SA(13) 
    BR    14  
STATIC_DATA DS 0D
    IHASAVER DSECT=YES,LIST=YES,SAVER=YES,SAVF4SA=COND
wto_b DSECT
         DS    0H
wto_len  DS    AL2
wto_flag DS    AL2
wto_msg  DS    CL252
wto_size EQU *-wto_b
    END
