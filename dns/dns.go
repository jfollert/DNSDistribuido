package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
	"log"
	"net"
	"context"

	pb "../proto"
	"google.golang.org/grpc"
)

// ESTRUCTURAS
type Server struct{}

type RegistroZF struct{
	ruta string  // ruta dentro del sistema donde se almacena el archivo de Registro ZF
	rutaLog string // ruta dentro del sistema donde se almacena el archivo de Logs de Cambios.
	reloj [3]int
	dominioLinea map[string]int // relaciona el nombre de dominio a la linea que ocupa dentro del archivo de registro
}

type NodeInfo struct {
	Id   string `json:"id"`
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

type Config struct {
	DataNode []NodeInfo `json:"DataNode"`
	NameNode NodeInfo   `json:"NameNode"`
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

// FUNCIONES DEL SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}


func main() {
	log.Printf("= INICIANDO DNS SERVER =")

	// Definir e inicializar variables
	log.Printf("Inicializando variables")
	var dominioRegistro map[string]RegistroZF // relaciona el nombre de dominio con su Registro ZF respectivo
	dominioRegistro = make(map[string]RegistroZF)
	log.Printf("%v", dominioRegistro)
	
	// Cargar archivo de configuración
	log.Printf("Cargando archivo de configuración")
	var config Config
	config = cargarConfig("config.json")
	log.Printf("Archivo de configuración cargado")

	port := config.NameNode.Port

	iniciarNodo(port)

}