       IDENTIFICATION DIVISION.
       PROGRAM-ID. PFIX.
       AUTHOR. who.

       DATA DIVISION.
       WORKING-STORAGE SECTION.
       01 HIGH-PRECISION-NUMBER PIC S9(10)V9(8) COMP-3.
       01  STACK.
           05 ITEM OCCURS 100 TIMES.
              10 ITEM-VALUE PIC S9(10)V9(8).
       01  STACKPTR          PIC 9(3) VALUE 1.

       01  CF-INPUT-STRING   PIC X(20).
       01  CF-NUM            PIC S9(10)V9(8).
       01  CF-NUM1           PIC S9(10)V9(8).
       01  CF-NUM2           PIC S9(10)V9(8).
       01  CF-OPERATOR       PIC X(1).
       01  CF-RESULT         PIC S9(10)V9(8).
       01  CF-DISPLAY-RESULT PIC S9(10)V9(8).
       01  CF-VALID-INPUT    PIC X VALUE 'Y'.
       01  CF-INDEX          PIC 9(3).
       01  CF-TOP            PIC ---------9.99999999.
       01  CF-DEBUG          PIC X VALUE 'N'.

       PROCEDURE DIVISION.
       MAIN-PROCEDURE.
           PERFORM UNTIL CF-INPUT-STRING = 'quit'
               IF CF-DEBUG = 'Y'
                   DISPLAY "#>"
               END-IF
               ACCEPT CF-INPUT-STRING
               PERFORM VALIDATE-INPUT
               IF CF-VALID-INPUT = 'Y'
                   PERFORM PROCESS-INPUT
               END-IF
           END-PERFORM.
           STOP RUN.

       VALIDATE-INPUT.
           MOVE 'Y' TO CF-VALID-INPUT.
           EVALUATE TRUE
               WHEN CF-INPUT-STRING = 'stacktop'
                   CONTINUE
               WHEN CF-INPUT-STRING = 'quit'
                   CONTINUE
               WHEN CF-INPUT-STRING = 'clear'
                   CONTINUE
               WHEN CF-INPUT-STRING = 'stackindex'
                   CONTINUE
               WHEN OTHER
                   IF CF-INPUT-STRING = '+' OR CF-INPUT-STRING = '-' OR
                      CF-INPUT-STRING = '*' OR CF-INPUT-STRING = '/'
                       CONTINUE
                   ELSE
                       PERFORM CHECK-NUMERIC
                   END-IF
           END-EVALUATE.

       CHECK-NUMERIC.
           MOVE 0 TO CF-NUM.
           COMPUTE CF-NUM = FUNCTION NUMVAL(CF-INPUT-STRING).
           IF CF-NUM = 0 AND CF-INPUT-STRING NOT = '0' AND
              CF-INPUT-STRING NOT = '0.0'
               DISPLAY "# Error: Invalid input " CF-INPUT-STRING
               MOVE 'N' TO CF-VALID-INPUT
           END-IF.

       PROCESS-INPUT.
           EVALUATE TRUE
               WHEN CF-INPUT-STRING = 'stacktop'
                   PERFORM DUMP-STACK
                   CONTINUE
               WHEN CF-INPUT-STRING = 'stackindex'
                   DISPLAY STACKPTR
                   CONTINUE
               WHEN CF-INPUT-STRING = 'clear'
                   MOVE 1 to STACKPTR
                   MOVE 0 to CF-DISPLAY-RESULT 
                   CONTINUE
               WHEN CF-INPUT-STRING = 'quit'
                   CONTINUE
               WHEN CF-INPUT-STRING = '+' OR CF-INPUT-STRING = '-' OR
                   CF-INPUT-STRING = '*' OR CF-INPUT-STRING = '/'
                   PERFORM PERFORM-OPERATION
               WHEN OTHER
                   PERFORM PUSH-TO-STACK
           END-EVALUATE.
           IF CF-DEBUG = 'Y'
               PERFORM DUMP-STACK
           END-IF.

       PERFORM-OPERATION.
           IF STACKPTR < 3
               DISPLAY "# Error: Not enough operands for operation"
           ELSE
               PERFORM POP1-FROM-STACK
               PERFORM POP2-FROM-STACK
               EVALUATE CF-INPUT-STRING
                   WHEN "+"
                       COMPUTE CF-RESULT = CF-NUM2 + CF-NUM1
                   WHEN "-"
                       COMPUTE CF-RESULT = CF-NUM2 - CF-NUM1
                   WHEN "*"
                       COMPUTE CF-RESULT = CF-NUM2 * CF-NUM1
                   WHEN "/"
                       IF CF-NUM1 = 0
                           DISPLAY "# Error: Division by zero"
                           PERFORM PUSH-TO-STACK *> Push CF-NUM2 back
                           PERFORM PUSH-TO-STACK *> Push CF-NUM1 back
                           EXIT PARAGRAPH
                       ELSE
                           COMPUTE CF-RESULT = CF-NUM2 / CF-NUM1
                       END-IF
               END-EVALUATE
               PERFORM PUSH-TO-STACK
           END-IF.

       POP1-FROM-STACK.
           SUBTRACT 1 FROM STACKPTR.
           MOVE ITEM(STACKPTR) TO CF-NUM1.

       POP2-FROM-STACK.
           SUBTRACT 1 FROM STACKPTR.
           MOVE ITEM(STACKPTR) TO CF-NUM2.

       PUSH-TO-STACK.
           IF STACKPTR > 100
               DISPLAY "# Error: Stack overflow"
           ELSE
               IF CF-INPUT-STRING = '+' OR CF-INPUT-STRING = '-' OR
                  CF-INPUT-STRING = '*' OR CF-INPUT-STRING = '/'
                   MOVE CF-RESULT TO ITEM(STACKPTR)
               ELSE
                   MOVE CF-NUM TO ITEM(STACKPTR)
               END-IF
               ADD 1 TO STACKPTR
           END-IF.

       DUMP-STACK.
               IF CF-DEBUG = 'Y'
                   DISPLAY "# Stack: "
               END-IF
               PERFORM VARYING CF-INDEX FROM 1 BY 1
                       UNTIL CF-INDEX >= STACKPTR
                   MOVE ITEM(CF-INDEX) TO CF-DISPLAY-RESULT
                   IF CF-DEBUG = 'Y'
                       DISPLAY "#" CF-DISPLAY-RESULT
                   END-IF
               END-PERFORM
               MOVE CF-DISPLAY-RESULT TO CF-TOP
               DISPLAY CF-TOP.
