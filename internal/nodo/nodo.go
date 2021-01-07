package nodo

import (
	"log"
	"context"
	"errors"
	"time"
	//"net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	pb "github.com/jfomu/DNSDistribuido/internal/proto"
)

type Server struct{}

func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Consulta) (*pb.Estado, error){
	return nil, errors.New("Función ObtenerRegistro() no implementada para este nodo.")
}

func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	return nil, errors.New("Función Get() no implementada para este nodo.")
}

func (s *Server) Create(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	return nil, errors.New("Función Create() no implementada para este nodo.")
}

func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	return nil, errors.New("Función Delete() no implementada para este nodo.")
}

func (s *Server) Update(ctx context.Context, message *pb.ConsultaUpdate) (*pb.RespuestaAdmin, error){
	return nil, errors.New("Función Update() no implementada para este nodo.")
}

func (s *Server) GetFile(message *pb.Consulta, srv pb.ServicioNodo_GetFileServer) error{
	return errors.New("Función GetFile() no implementada para este nodo.")
}

func (s *Server) SetFile(stream pb.ServicioNodo_SetFileServer) error{
	return errors.New("Función SetFile() no implementada para este nodo.")
}

func (s *Server) GetDominios(ctx context.Context, message *pb.Vacio) (*pb.Dominios, error){
	return nil, errors.New("Función GetDominios() no implementada para este nodo.")
}


/*
func IniciarNodo(port string) {
	// Iniciar servidor gRPC
	log.Printf("Iniciando servidor gRPC en el puerto " + port)
	lis, err := net.Listen("tcp", ":" + port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	//Registrar servicios en el servidor
	log.Printf("Registrando servicios en servidor gRPC\n")
	pb.RegisterServicioNodoServer(grpcServer, &servidor)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}

}
*/

func ConectarNodo(ip string, port string) (*grpc.ClientConn, error) {
	var conn *grpc.ClientConn
	log.Printf("Iniciando conexión con " + ip + ":" + port)
	host := ip + ":" + port
	keepConf:= keepalive.ClientParameters{
		Time:					10 * time.Second,
		Timeout:				2 * time.Second,
		PermitWithoutStream:	true,
	}
	conn, err := grpc.Dial(host, grpc.WithInsecure(), grpc.WithKeepaliveParams(keepConf))
	if err != nil {
		//log.Printf("No se pudo establecer la conexión con " + ip + ":" + strconv.Itoa(port))
		return nil, err
	}
	
	//log.Printf("Conexión establecida con " + ip + ":" + strconv.Itoa(port))
	return conn, nil
}