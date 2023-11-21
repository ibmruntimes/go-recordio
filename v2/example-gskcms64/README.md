An example of loading the GSKCMS64 dll (ssl API)

To build and run
```
go build 

./example-gskcms64
```

The output should be a list of CA Certificates on the system.

You might need to permit access like
```
RDEFINE CSFSERV profile-name UACC(NONE)
PERMIT profile-name CLASS(CSFSERV) ID(your-userid) ACCESS(READ)
```
