An example of using the 31-bit RACFAPI IRRSIM00 
You will likely encounter safrc 8 , racfrc 8 and reason code 20 (not authorized to used service)

To build
```
go build -o racfapi
```

To run
```
./racfapi
```

This is an example of loading and using an AMODE-31 module IRRSIM00 
https://www.ibm.com/docs/en/zos/2.5.0?topic=descriptions-r-usermap-irrsim00-map-application-user

All the arguments for AMODE-31 program must reside below the bar (< 2G address) 
1. We allocate a structure PLIST that contains all the arguments and return value in below the bar storage with utils.Malloc31()
2. We set up the arguments, conversion to EBCDIC is neccessary for character arguments.
3. We set up a standard OS parameter list of an array of 4-byte pointers to the arguments, and turn on the first bit of the address for the last argument.
4. Use utils.LoadMod() to load IRRSIM00
5. Invoke IRRSIM00 with mod.Call
6. Check return code and return results
8. Unload module
9. Free below the bar storage.
