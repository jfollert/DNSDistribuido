## EJECUCIÓN DE NODOS
cliente:
	go run cmd/cliente/cliente.go

dns:
	go run cmd/dns/dns.go

admin:
	go run cmd/admin/admin.go

broker:
	go run cmd/broker/broker.go


clean:
	rm -rf logs/*
	rm -rf registros/*


vm:
	mv config.json config_local.vm
	mv config_vm.json config.json

local:
	mv config.json config_vm.json
	mv config_local.json config.json

#protoc:
#	export PATH="$PATH:$(go env GOPATH)/bin"
#	protoc -I proto --go_out=plugins=grpc:proto proto/*.proto

