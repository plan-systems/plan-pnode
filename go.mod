module github.com/plan-systems/plan-pnode

go 1.12

require (
	github.com/plan-systems/plan-core v0.0.3
	google.golang.org/grpc v1.22.0
)

replace github.com/plan-systems/plan-core => ../plan-core

replace github.com/plan-systems/plan-pdi-local => ../plan-pdi-local
