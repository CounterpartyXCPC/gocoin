# About Gocoin-cash

**Gocoin-cash** is a full **Bitcoin Cash** solution written in Go language (golang) and is based on 
the original work of of Gocoin (https://github.com/piotrnar/gocoin) by Piotr Narewski.

The software architecture is focused on maximum performance of the node
\and cold storage security of the wallet.

The **client** (p2p node) is an application independent from the **wallet**.
It keeps the entire UTXO set in RAM, providing the best block processing performance on the market.

In it's original form, With a decent machine and a fast connection (e.g. 4 vCPUs from Google Cloud or 
Amazon AWS), the node would sync the entire bitcoin block chain in less than 4 hours (as of chain 
height ~512000) for the Bitcoin core chain.

Benchmarks for the Bitcoin BCH chain to follow.

The **wallet** is designed to be used offline.

It is deterministic and password seeded.

As long as you remember the password, you do not need any backups ever.

# Requirements

## Hardware

**client**:

* 64-bit architecture OS and Go compiler.
* File system supporting files larger than 4GB.
* At least 15GB of system memory (RAM).

**wallet**:

* Any platform that you can make your Go (cross)compiler to build for (Raspberry Pi works).
* For security reasons make sure to use encrypted swap file (if there is a swap file).
* If you decide to store your password in a file, have the disk encrypted (in case it gets stolen).

## Operating System
Having hardware requirements met, any target OS supported by your Go compiler will do.
Currently that can be at least one of the following:

* Windows
* Linux
* OS X
* Free BSD

## Build environment
In order to build Gocoin-cash yourself, you will need the following tools installed in your system:

* **Go** (version 1.8 or higher) - http://golang.org/doc/install
* **Git** - http://git-scm.com/downloads

If the tools mentioned above are all properly installed, you should be able to execute `go` and `git`
from your OS's command prompt without a need to specify full path to the executables.

### Linux

When building for Linux make sure to have `gcc` installed or delete file `lib/utxo/membind_linux.go`

# Getting sources

Use `go get` to fetch and install the source code files.
Note that source files get installed within your GOPATH folder.

	go get github.com/CounterpartyXCPC/gocoin-cash

# Building

## Client node
Go to the `client/` folder and execute `go build` there.

## Wallet
Go to the `wallet/` folder and execute `go build` there.

## Tools
Go to the `tools/` folder and execute:

	go build btcversig.go

Repeat the `go build` for each source file of the tool you want to build.

# Binaries

Windows or Linux (amd64) binaries can be downloaded from

(T.B.A) @todo, No binaries yet released (Fri Jun 15, 2018)

Please note that the binaries are usually not up to date.
We strongly encourage everyone to build the binaries.

# Development
Although it is an open source project, we will accept merge and any pull requests, however we need to
have contributors assign copyright to the CCA so we may continue to operate within the bounds of our
limited software license agreement to the original project. The reason is that Piotr Narewsk retains 
certain rights as the original author of this software and excersises an interest and control over its
licensing. The CCA accepts these terms without issue and supports the author's rights in this regard.

# Support
The official web page of the project is served at <a href="http://Gocoin-cash.com">Gocoin-cash.com</a>
where you can find extended documentation, including a **User Manual**.

Please log github issues here when you have questions concerning this software.
