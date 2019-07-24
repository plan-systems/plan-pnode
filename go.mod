module github.com/plan-systems/plan-pnode

go 1.12

require (
	github.com/dgraph-io/badger v1.6.0 // indirect
	github.com/plan-systems/klog v0.0.0-20190618231738-14c6677fa6ea // indirect
	github.com/plan-systems/plan-core v0.0.0-20190714185042-db140cabb3fa
	github.com/plan-systems/plan-pdi-local v0.0.0-20190713092621-7d241122cbd5 // indirect
	google.golang.org/grpc v1.22.0
)

// replace github.com/plan-systems/plan-core => ../plan-core
