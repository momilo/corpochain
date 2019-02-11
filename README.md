# Corpochain
[![Build Status](https://travis-ci.org/joemccann/dillinger.svg?branch=master)](https://travis-ci.org/joemccann/dillinger)

Corpochain is a blockchain based proof of work system for financial data storage for large organizations.

Financial data requires secure storage mechanism, corpochain offers distributed solution for large datasets.

Corpochain benefits:
- Efficient Transactional data storage
- Workflow and execution analysis
- Life data storage cost anaylysis
- Configuration


#Installation
- Install `Go 1.11.` Earlier versions won’t work.
- Install `tmux`
- `git clone` the code into `$GOPATH`, typically `~/go/src/corpochain/prototype`
- `make install` will install to `$GOPATH/bin/corpochain`. Make sure this location is on your path.


# Initialize a network
`corpochain run`

# Run with docker
`docker-compose up --build`

Gateway node available at port:
`:8080`

# Commands

Full documentation on API will be released soon.


