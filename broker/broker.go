package main

import (
	"fmt"
	"log"
	"io/ioutil"
	"context"
	"os"
	"encoding/json"
	"net"
	"strings"
	//"math/rand"

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
	DNS[]NodeInfo `json:"DNS"`
	Broker NodeInfo `json:"Broker"`
}

var config Config

// FUNCIONES DEL SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}

func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	respuesta := new(pb.Respuesta)
	if strings.Compare("", message.NombreDominio) == 0 { // Si no se recibe un nombreDominio
		log.Printf("Recibida solicitud desde administrador, buscando servidor DNS")
		//idRandom := rand.Intn(3)
		idRandom := 0
		//log.Printf("DNS: %v", config.DNS)
		log.Printf("Servidor DNS obtenido de forma aleatoria: DNS%d", idRandom+1)
		respuesta.Ip = config.DNS[idRandom].Ip
		respuesta.Port = config.DNS[idRandom].Port
	}
	return respuesta, nil
}

func (s *Server) Create(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	return new(pb.RespuestaAdmin), nil
}

func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	return new(pb.RespuestaAdmin), nil
}


func (s *Server) Update(ctx context.Context, message *pb.ConsultaUpdate) (*pb.RespuestaAdmin, error){
	return new(pb.RespuestaAdmin), nil
}


// FUNCIONES
func cargarConfig(file string) {
    log.Printf("Cargando archivo de configuración")
    configFile, err := ioutil.ReadFile(file)
    if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	json.Unmarshal(configFile, &config)
	log.Printf("Archivo de configuración cargado")
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

	// Cargar archivo de configuración
	cargarConfig("config.json")

	iniciarNodo("9000")
}