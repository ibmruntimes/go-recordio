To run GO program as BATCH, the GO program need to be copied to a PDSE

e.g.

go build batch1.go

cp -X batch1 "//'YOUR.GO.PDSE(BATCH1)'"

gobatch.jcl constains a sample JCL.

The batch1 example copies SYSIN to SYSPRINT
