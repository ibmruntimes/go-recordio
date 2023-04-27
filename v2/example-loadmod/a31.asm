         title 'amode31'
main     csect ,
main     amode 31
main     rmode any
         stm      14,12,12(13)
         lr       12,15
         using    main,12
         la       15,0
         l        0,write_sz
         getmain  RU,LV=(0),SP=(15)
         lr       11,1
         using    writable,11
         st       13,4(,11)
         st       11,8(,13)
         lm       15,1,16(13)
         lr       13,11
ptr      equ      0
wto_str  equ      0
wto_len  equ      0
strlen_parm equ   0
change   equ      0
arg      equ      0
parm     equ      0
parm_len equ      parm
parm_str equ      parm+2
string   equ      0
message  equ      0
         lr       2,1
         st       2,plist
         mvhi     last_parm,1
         mvhhi    wtob_flag,-32768
         mvhhi    wtob_sz,64
         mvc      wtob_msgarea(60),sepline
         mvhhi    wtob_msgarea+60,512
         mvhhi    wtob_msgarea+62,32
         la       1,wtob
         svc      35
while_not_last ds 0h
         l        5,arg(,2)
         st       5,last_parm
         llh      1,parm_len(,5)
         chi      1,128
         jnl      not_vchr
         lr       3,1
         mvhhi    wtob_flag,-32768
         ahik     0,1,4
         ahi      3,-1
         sth      0,wtob_sz
         ex       3,mvc_tpl_2
         la       1,wtob_msgarea(1)
         mvhhi    0(1),512
         mvhhi    2(1),32
         la       1,wtob
         svc      35
         j        next_1
not_vchr l        4,arg(,2)
         lhi      3,0
         mvhi     result,0
         cli      string(4),0
         jnh      fnd_null
while_not_null ds 0h
         alfi     3,x'00000001'
         lhi      0,1
         st       3,result
         alr      4,0
         cli      string(4),0
         jh       while_not_null
fnd_null st       3,wrk_2
         st       3,tmplen
         l        0,arg(,2)
         st       0,wrk_1
         st       3,wrk_3
         chi      3,252
         mvhhi    wtob_flag,-32768
         jnh      tmp_l1
         mvhi     wrk_3,252
tmp_l1   l        1,wrk_3
         ahik     0,1,4
         sth      0,wtob_sz
         l        6,wrk_3
         l        4,wrk_1
         ahi      6,-1
         exrl     6,mvc_tpl_1
         la       1,wtob_msgarea(1)
         mvhhi    0(1),512
         mvhhi    2(1),32
         la       1,wtob
         svc      35
         chi      3,12
         jnh      next_1
         l        4,arg(,2)
         mvc      change(12,4),testdata
next_1   alfi     2,x'00000004'
         st       2,plist
         mvhhi    wtob_flag,-32768
         mvhhi    wtob_sz,64
         mvc      wtob_msgarea(60),sepline
         mvhhi    wtob_msgarea+60,512
         mvhhi    wtob_msgarea+62,32
         la       1,wtob
         svc      35
         ltr      5,5
         jp       while_not_last
         lr       1,11
         l        13,4(,13)
         la       15,0
         l        0,write_sz
         freemain RU,LV=(0),A=(1),SP=(15)
         lhi      15,0
         l        14,12(,13)
         lm       0,12,20(13)
         bcr      15,14
         ltorg
write_sz dc       a(write_area_size)
mvc_tpl_1 mvc     wtob+4(0),message(4)
mvc_tpl_2 mvc     wtob+4(0),message+2(5)
sepline  dc       c'------------------------------'
         dc       c'------------------------------'
testdata dc       cl12'-overwrote- '
writable dsect
         ds       18f
wrk_3    ds       f
last_parm ds      f
result   ds       f
wtob     ds       cl260
         org      wtob
wtob_sz  ds       hl2
wtob_flag ds      hl2
wtob_msgarea ds   cl256
wrk_1    ds       f
wrk_2    ds       f
plist    ds       a
tmplen   ds       f
         org      *+1-(*-writable)/(*-writable)
endmark ds       0x
write_area_size equ ((endmark-writable+7)/8)*8
         end
