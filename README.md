# Self-Upgrade

A server Golang implementation that enables to `Self-Upgrade` with zero downtime.  
The solution uses only the Go standard libraries.

## Quick start
```
 git clone go get github.com/slayer/autorestart
 
 go build
 ./self-upgrade 

```

After start, the web application is available at:
```
http://localhost:9000
```

#### Features
- Self-Upgrade
- Verify executable RSA signaure
- Zero Downtime restart

#### Implementation details

A example of updated executable file is stored on the `/dist` directory. We assume that the version extension is specified by `.verX` and an aded extension `.RSAsignature` for their associated RSA signature.

#### Limitations/TODO

- Command Line Argument
- Generic file naming
- Not running on Windows








