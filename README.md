## Decentralized VPN
[![Build Status](https://travis-ci.org/mehrdadrad/radvpn.svg?branch=master)](https://travis-ci.org/mehrdadrad/radvpn) 
[![Go Report Card](https://goreportcard.com/badge/github.com/mehrdadrad/radvpn)](https://goreportcard.com/report/github.com/mehrdadrad/radvpn)
[![GoDoc](https://godoc.org/github.com/mehrdadrad/radvpn?status.svg)](https://godoc.org/github.com/mehrdadrad/radvpn)

![Alt text](/docs/imgs/radvpn.png?raw=true "radvpn")

## Docker
```
docker run --privileged -d -p 8085:8085 -v $(pwd)/radvpn.yaml:/etc/radvpn.yaml -e RADVPN_NODE_NAME=node1 radvp
```

## License
Licensed under the Apache License, Version 2.0 (the "License")

## Contribute
Welcomes any kind of contribution, please follow the next steps:

- Fork the project on github.com.
- Create a new branch.
- Commit changes to the new branch.
- Send a pull request.

