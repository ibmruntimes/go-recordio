An example of using the 31-bit module IEANTRT to determine if syslodd is up

To build
```
go build -o ieantrt
```

To run
```
./ieantrt
```


This is an example of loading and using an AMODE-31 module IEANTRT to determine if the service syslogd is up and runing.
All the arguments for AMODE-31 program must reside below the bar (< 2G address) 
1. We allocate a structure PLIST that contains all the arguments and return value in below the bar storage with utils.Malloc31()
2. We set up the arguments, conversion to EBCDIC is neccessary for character arguments.
3. We set up a standard OS parameter list of an array of 4-byte pointers to the arguments, and turn on the first bit of the address for the last argument.
4. Use utils.LoadMod() to load IEANTRT
5. Invoke IEANTRT with mod.Call
6. Check return code
7. Determine whether syslogd is active.
8. Unload module
9. Free below the bar storage.
10. Print result.

