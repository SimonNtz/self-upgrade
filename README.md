# Self-Upgrade

An example of Golang server that enables to `Self-Upgrade` itself with zero downtime.  
The solution only uses the Go standard libraries.

## Quick start
```
 git clone https://github.com/SimonNtz/self-upgrade.git
 
 go build
 ./self-upgrade 

```
The web application is available at:

```
http://localhost:9000
```

The executable's version displayed by the front-end template is hardcoded into the `server.go` file.
The executables can be signed and verified using the tests in `verifier_test.go`. You can run them with the following command:

`go test`


*note: After upgrading and testing, the main executable is replaced. Therefore do not forget to build the source code again if you want to redo the demo.


#### Features
- Self-Upgrade
- Verify executable RSA signaure
- Zero Downtime restart

#### Implementation details

A example of updated executable file is stored on the `dist` directory. We assume that its version is specified by the `.verX` name extension. Its associated RSA signature file - located in the same folder - have the `.RSAsignature` added extension.

***

#### Limitations

- Not running on Windows

### TODO

- Pass the executable version using `go build -ldflags "-X main.Version"
- Enable generic file naming








