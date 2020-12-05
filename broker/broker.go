package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"context"
	"os"
	"encoding/json"
	"net"

	pb "../proto"
	"google.golang.org/grpc"
)

// ESTRUCTURAS
type Server struct{}

type NodeInfo struct {
	Id string `json:"id"`
	Ip string `json:"ip"`
	Port string `json:"port"`
}

type Config struct {
	DataNode []NodeInfo `json:"DataNode"`
	NameNode NodeInfo `json:"NameNode"`
}

// FUNCIONES DEL SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}


// FUNCIONES
func cargarConfig(file string) Config {
    var config Config
    configFile, err := ioutil.ReadFile(file)
    if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(configFile, &config)
    return config
}

func iniciarNodo(port string) {
	// Iniciar servidor gRPC
	log.Printf("Iniciando servidor gRPC en el puerto " + port)
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := Server{}
	grpcServer := grpc.NewServer()

	//Registrar servicios en el servidor
	log.Printf("Registrando servicios en servidor gRPC\n")
	pb.RegisterServicioNodoServer(grpcServer, &s)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

}

func main() {
	log.Println("= INICIANDO BROKER =")
	iniciarNodo("9000")
}