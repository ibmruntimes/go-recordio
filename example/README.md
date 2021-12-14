Example 1:

test.go

```
go build test.go
```

VSAM created by using the crtvsamxsysvar script from the source: 

```
git@github.com:MikeFultonDev/samples.git, branch Xsysvar.

```
Note: to use this you also need Z Open Automation Utilities (ZOAU) installed.


Example 2:

testpds.go

```
go build testpds.go
```


Reading the PDS member SYS1.MACLIB(EXCP) and print it on the screen


Example 3:

batch1.go

```
go build batch1.go

cp -X batch1 "//'YOUR.GO.PDSE(BATCH1)'"
```

gobatch.jcl constains a sample JCL which you have to modify.

This example copies SYSIN to SYSPRINT


Example 4:

batch2.go

```
go build batch2.go

cp -X batch2 "//'YOUR.GO.PDSE(BATCH2)'"
```

gobatch.jcl constains a sample JCL which you have to modify.

This example sends GO print() etc output to SYSPRINT

** Note that BATCH mode is not officially supported, it is only 
used to demonstrate this package.

