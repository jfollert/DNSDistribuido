Cliente:
	go run cliente/cliente.go

DNS:
	go run dns/dns.go

Admin:
	go run admin/admin.go

Broker:
	go run broker/broker.go

Protoc:
	export PATH="$PATH:$(go env GOPATH)/bin"
	protoc -I proto --go_out=plugins=grpc:proto proto/*.proto

clean:
	rm -rf dns/logs/*.log
	rm -rf dns/registros/*.zf