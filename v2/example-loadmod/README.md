An example of using load modules

To build and run
```
./dotest
```

This is an example of calling user assembly programs of different amodes.
The assembly programs are very simple and just print some information to the system console via WTO.
The output from these programs are captured with setenv _BPXK_JOGLOG=fd 
in which fd is a pipe to write into. Exececution of amode24, amode31 and amod64 programs are very similar, the noticible differences are 
amode24 program's arguments and returns are allocated with utils.Malloc24, amode31 program's arguments and returns are allocated with utils.Malloc31.


