CBL OPT(FULL),LIST
       ID DIVISION.
       PROGRAM-ID. 'XDUMP2C' recursive.
       ENVIRONMENT DIVISION.
       INPUT-OUTPUT Section.
       DATA DIVISION.
       FILE SECTION.

       Working-Storage Section.

       01 aprint.
         02 agroup.
           05 a0 pic X(16) VALUE '................'.  
           05 a1 pic X(16) VALUE '................'. 
           05 a2 pic X(16) VALUE ' !"#$%&''()*+,-./'. 
           05 a3 pic X(16) VALUE '0123456789:;<=>?'. 
           05 a4 pic X(16) VALUE '@ABCDEFGHIJKLMNO'. 
           05 a5 pic X(16) VALUE 'PQRSTUVWXYZ[\]^_'. 
           05 a6 pic X(16) VALUE '`abcdefghijklmno'. 
           05 a7 pic X(16) VALUE 'pqrstuvwxyz{|}~.'. 
           05 a8 pic X(16) VALUE '................'. 
           05 a9 pic X(16) VALUE '................'. 
           05 aa pic X(16) VALUE '................'. 
           05 ab pic X(16) VALUE '................'. 
           05 ac pic X(16) VALUE '................'. 
           05 ad pic X(16) VALUE '................'. 
           05 ae pic X(16) VALUE '................'. 
           05 af pic X(16) VALUE '................'. 
         02 atable redefines agroup pic x(256).

       01 eprint.
         02 egroup.
           05 e0 pic X(16) VALUE '................'.
           05 e1 pic X(16) VALUE '................'.
           05 e2 pic X(16) VALUE '................'.
           05 e3 pic X(16) VALUE '................'.
           05 e4 pic X(16) VALUE ' ...........<(+|'.
           05 e5 pic X(16) VALUE '&.........!$*);^'.
           05 e6 pic X(16) VALUE '-/.........,%_>?'.
           05 e7 pic X(16) VALUE '.........`:#@''="'.
           05 e8 pic X(16) VALUE '.abcdefghi......'.
           05 e9 pic X(16) VALUE '.jklmnopqr......'.
           05 ea pic X(16) VALUE '.~stuvwxyz...[..'.
           05 eb pic X(16) VALUE '.............]..'.
           05 ec pic X(16) VALUE '{ABCDEFGHI......'.
           05 ed pic X(16) VALUE '}JKLMNOPQR......'.
           05 ee pic X(16) VALUE '\.STUVWXYZ......'.
           05 ef pic X(16) VALUE '0123456789......'.
         02 etable redefines egroup pic x(256).

       01 hprint.
         02 hgroup.
           05 h0 pic X(32) VALUE '000102030405060708090a0b0c0d0e0f'.
           05 h1 pic X(32) VALUE '101112131415161718191a1b1c1d1e1f'.
           05 h2 pic X(32) VALUE '202122232425262728292a2b2c2d2e2f'.
           05 h3 pic X(32) VALUE '303132333435363738393a3b3c3d3e3f'.
           05 h4 pic X(32) VALUE '404142434445464748494a4b4c4d4e4f'.
           05 h5 pic X(32) VALUE '505152535455565758595a5b5c5d5e5f'.
           05 h6 pic X(32) VALUE '606162636465666768696a6b6c6d6e6f'.
           05 h7 pic X(32) VALUE '707172737475767778797a7b7c7d7e7f'.
           05 h8 pic X(32) VALUE '808182838485868788898a8b8c8d8e8f'.
           05 h9 pic X(32) VALUE '909192939495969798999a9b9c9d9e9f'.
           05 ha pic X(32) VALUE 'a0a1a2a3a4a5a6a7a8a9aaabacadaeaf'.
           05 hb pic X(32) VALUE 'b0b1b2b3b4b5b6b7b8b9babbbcbdbebf'.
           05 hc pic X(32) VALUE 'c0c1c2c3c4c5c6c7c8c9cacbcccdcecf'.
           05 hd pic X(32) VALUE 'd0d1d2d3d4d5d6d7d8d9dadbdcdddedf'.
           05 he pic X(32) VALUE 'e0e1e2e3e4e5e6e7e8e9eaebecedeeef'.
           05 hf pic X(32) VALUE 'f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff'.
         02 xtable redefines hgroup pic x(512).

       Local-Storage Section.

       01 i      pic s9(9) usage is comp-5.
       01 j      pic s9(9) usage is comp-5.
       01 l      pic s9(9) usage is comp-5.
       01 dummy  pic s9(9) usage is comp-5.
       01 rem    pic s9(9) usage is comp-5.
       01 rcnt   pic s9(9) usage is comp-5.
       01 aarea  pic x(16).
       01 earea  pic x(16).
       01 xarea  pic x(32).
       01 xarea2 pic x(32).
      * big endian only
       01 convchar2int.
          05 chargroup.
             10 zz  pic x value x'00'.
             10 num1byte pic x.
          05 nval redefines chargroup pic s9(2) comp-5.

       Linkage Section.

       01 p      usage is pointer.
       01 fp     usage is function-pointer.
       01 cnt    pic 9(9) usage is comp-5.
       01 dat    pic x(65536).

       Procedure Division using by value fp by value p by value cnt.
       Begin. 
           display 'offset_____ 0_______ 4_______ 8_______ 12______ ',
                            'ASCII___________ ',  
                            'EBCDIC__________ '.
           move 0 to rcnt.
           move "@@@" to xarea2.
           set address of dat to p;
           move 0 to i.
           move i to l;
           perform with test after until i = cnt
             divide i by 16 giving dummy remainder rem end-divide
             move dat( 1 + i : 1) to num1byte
             move nval to j
             move atable( 1 + j : 1) to aarea( 1 + rem : 1) 
             move etable( 1 + j : 1) to earea( 1 + rem : 1) 
             move xtable( 1 + (j * 2) : 2 ) to xarea( 1 + (rem * 2) : 2)
             if (i > 0 ) and (rem = 15) then 
                if (xarea equal to xarea2) then
                   add 1 to rcnt
                else 
                   if rcnt greater than 0 then 
                     display ". . . . . . . . . >> " rcnt 
                             " lines same as above"
                     move 0 to rcnt
                   end-if
                   display l, ': ', xarea(1:8), ' ',
                        xarea(9:8), ' ',
                        xarea(17:8), ' ',
                        xarea(25:8), ' ',
                        aarea, ' ', earea
                end-if
                move xarea to xarea2
                add 16 to l;
             end-if 
             add 1 to i
           end-perform.
           if rcnt greater than 0 then 
             display ". . . . . . . . . >> " rcnt 
                     " lines same as above"
             move 0 to rcnt
           end-if
           if (rem not equal 15) then 
                initialize aarea( rem + 2:)
                initialize earea( rem + 2:)
                initialize xarea( 1 + (rem + 1) * 2 :)
                display l, ': ', xarea(1:8), ' ',
                        xarea(9:8), ' ',
                        xarea(17:8), ' ',
                        xarea(25:8), ' ',
                        aarea, ' ', earea
           end-if.
           CALL fp USING BY value p, cnt.
           goback.
       END PROGRAM 'XDUMP2C'.
