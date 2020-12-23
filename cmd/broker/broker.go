package main

import (
	"log"
	"context"
	"net"
	"math/rand"
	"errors"

	pb "github.com/jfomu/DNSDistribuido/internal/proto"
	"github.com/jfomu/DNSDistribuido/internal/config"
	"google.golang.org/grpc"
)

//// ESTRUCTURAS
type Server struct{}

//// VARIABLES GLOBALES
var configuracion *config.Config

//// FUNCIONES
func dnsAleatorio() (string, string){
	idRandom := rand.Intn(3)
	//idRandom := 0
	log.Printf("Servidor DNS obtenido de forma aleatoria: DNS%d\n", idRandom+1)
	return configuracion.DNS[idRandom].Ip, configuracion.DNS[idRandom].Port
}


//// FUNCIONES DEL SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}

func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	var dnsIP string
	var dnsPort string
	var respuesta *pb.Respuesta

	// Seleccionar servidor DNS
	if message.Ip != "" && message.Port != "" {  // Si se recibieron IP y puerto como argumentos
		dnsIP = message.Ip
		dnsPort = message.Port
	} else { // Si se debe entregar un servidor aleatorio
		dnsIP, dnsPort = dnsAleatorio()
	}

	if message.Ip == "" && message.Port == "" && message.NombreDominio == "" { // Se es una consulta del administrador
		log.Println("Enviando DNS aleatorio al Admin")
		respuesta = &pb.Respuesta{Ip: dnsIP, Port: dnsPort}
		return respuesta, nil
	}

	conn, err := conectarNodo(dnsIP, dnsPort)
	if err != nil{
		log.Printf("Error al intentar realizar conexión gRPC: %s\n", err)
		return nil, err
	}

	dnsServer := pb.NewServicioNodoClient(conn)
	respuesta, err = dnsServer.Get(context.Background(), message)
	if err != nil{
		log.Printf("Error al intentar conectar al servidor del servicio: %s\n", err)
		return nil, err
	}

	return respuesta, nil
}

func (s *Server) Create(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	return nil, errors.New("Método Create() no implementado en este nodo")
}

func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	return nil, errors.New("Método Delete() no implementado en este nodo")
}


func (s *Server) Update(ctx context.Context, message *pb.ConsultaUpdate) (*pb.RespuestaAdmin, error){
	return nil, errors.New("Método Update() no implementado en este nodo")
}


// FUNCIONES
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

func conectarNodo(ip string, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	log.Printf("Intentando iniciar conexión con " + ip + ":" + port)
	host := ip + ":" + port
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		//log.Printf("No se pudo establecer la conexión con " + ip + ":" + strconv.Itoa(port))
		return nil, err
	}
	//log.Printf("Conexión establecida con " + ip + ":" + strconv.Itoa(port))
	return conn, nil
}

func main() {
	log.Println("= INICIANDO BROKER =")

	// Cargar archivo de configuración
	configuracion = config.GenConfig("config.json")

	iniciarNodo(configuracion.Broker.Port)
}