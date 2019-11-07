## Decentralized VPN
[![Build Status](https://travis-ci.org/mehrdadrad/radvpn.svg?branch=master)](https://travis-ci.org/mehrdadrad/radvpn) 
[![Go Report Card](https://goreportcard.com/badge/github.com/mehrdadrad/radvpn)](https://goreportcard.com/report/github.com/mehrdadrad/radvpn)
[![GoDoc](https://godoc.org/github.com/mehrdadrad/radvpn?status.svg)](https://godoc.org/github.com/mehrdadrad/radvpn)

![Alt text](/docs/imgs/radvpn.png?raw=true "radvpn")

## Build
Given that the Go Language compiler (version 1.11 or greater is required) is installed, you can build it with:
```
go get https://github.com/mehrdadrad/radvpn
cd $GOPATH/src/github.com/mehrdadrad/radvpn
go build .
```

## Docker
```
docker run --privileged -d -p 8085:8085 -v $(pwd)/radvpn.yaml:/etc/radvpn.yaml -e RADVPN_NODE_NAME=node1 radvp
```

## License
This project is licensed under MIT license. Please read the LICENSE file.

## Contribute
Welcomes any kind of contribution, please follow the next steps:

- Fork the project on github.com.
- Create a new branch.
- Commit changes to the new branch.
- Send a pull request.

