An example of using the USS assembly interface

To build
```
go build -o pstree
```

To run
```
./pstree
```

This is an example of calling BPX4GTH

https://www.ibm.com/docs/en/zos/2.5.0?topic=64-bpx4gth-getthent-example

BPX4GTH is just one example out of the USS assembler callable services.
They can be accessed with utils.Bpxcall with different offsets.

All these services are available in amode-64,  we don't need to allocate below-the-bar storage for them nor need to allocate save area.

In this example, we use the Pgtha DSECT for input argument and the Pgthc, Pgthf DSECT for return data.

We obtain a list of processes with pid, ppid and arguments and draw a tree like diagram to standard output.

