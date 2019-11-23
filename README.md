# plan-pnode

```
         P urposeful
         L ogistics
         A rchitecture
P  L  A  N etwork
```

![](https://github.com/plan-systems/plan-pnode/workflows/Build%20and%20test/badge.svg)

[PLAN](http://plan-systems.org) is a free and open platform for groups to securely communicate, collaborate, and coordinate projects and activities.

## About

- This repo builds a daemon called `pnode`, the reference implementation of the [PLAN Data Model](https://github.com/plan-systems/design-docs/blob/master/PLAN-Proof-of-Correctness.md)
- `pnode` publishes the gRPC `Repo` service (defined in [repo.proto](https://github.com/plan-systems/plan-protobufs/blob/master/pkg/repo/repo.proto))
- A pnode initiates connections to PLAN PDI nodes via the gRPC `StorageProvder` service (defined in [pdi.proto](https://github.com/plan-systems/plan-protobufs/blob/master/pkg/pdi/pdi.proto))
- To better understand `pnode`, see the [Persistent Data Interface](https://github.com/plan-systems/design-docs/blob/master/PLAN-API-Documentation.md#Persistent-Data-Interface) docs and the PLAN Network Configuration Diagram


## Building

Requires golang 1.11 or above.

We're in the process of convering this project to use [go modules](https://github.com/golang/go/wiki/Modules). In the meantime, you'll want to checkout this repo into your `GOPATH` (or the default `~/go`).

```
mkdir -p ~/go/src/github.com/plan-systems
cd ~/go/src/github.com/plan-systems
git clone git@github.com:plan-systems/plan-pnode.git
cd plan-pnode
go get ./...
go build .
```
