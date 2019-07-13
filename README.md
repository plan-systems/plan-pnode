# plan-pnode

```
         P urposeful
         L ogistics
         A rchitecture
P  L  A  N etwork
```

[PLAN](http://plan-systems.org) is a free and open platform for groups to securely communicate, collaborate, and coordinate projects and activities.

## About

- This repo builds a daemon called `pnode`, the reference implementation of the [PLAN Data Model](https://github.com/plan-systems/design-docs/blob/master/PLAN-Proof-of-Correctness.md)
- `pnode` publishes the gRPC `Repo` service, defined in [repo.proto](https://github.com/plan-systems/plan-protobufs/blob/master/pkg/repo/repo.proto)
- A pnode _initiates_ connections to PLAN PDI nodes via the gRPC `StorageProvder` service, defined in [pdi.proto](https://github.com/plan-systems/plan-protobufs/blob/master/pkg/pdi/pdi.proto)
- To better understand `pnode`, see [Persistent Data Interface](https://github.com/plan-systems/design-docs/blob/master/PLAN-API-Documentation.md#Persistent-Data-Interface)

## Building

Requires golang 1.11 or above. This project uses [go modules](https://github.com/golang/go/wiki/Modules), although we're not yet pinning the `go.mod` and `go.sum` files until the upstream dependency [`plan-core`](https://github.com/plan-systems/plan-core) has stabilized. There's no need to set your `GOPATH` to build this project.

```
git clone git@github.com:plan-systems/plan-pnode.git
go test -v ./...
go build .
```
