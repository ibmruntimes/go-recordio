An example of using the MVS macro IEFSSREQ

To build
```
go build -o iefssreq
```

To run
```
./iefssreq
```

IEFSSREQ is a MVS macro described in
https://www.ibm.com/docs/en/zos/2.5.0?topic=subsystem-making-request-iefssreq-macro

This macro is a 31-bit macro and its entry point is acquired by navigating mvs control blocks.
In this example, it is implemented directly in Go assembly in utils/utils.s function IefssreqX
You can find out how to obtain the entry point by assembling IEFSSREQ with HLASM.
The Go assembly follows the same method. The extra wrap of SAM31 and SAM64 around BALR is for address mode switch to amode-31 and back.
This macro requires all its arguments to be in below-the-bar storage, so in this example we put them in a structure and acquire
the storage with utils.Malloc31. 
All character arguments are converted to EBCIDC with utils.AtoE
