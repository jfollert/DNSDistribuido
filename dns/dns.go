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
	DNS []NodeInfo `json:"DNS"`
	Broker NodeInfo   `json:"Broker"`
}

var config Config

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

func obtenerListaIPs() []string{
	var ips []string
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
					ip = v.IP
			case *net.IPAddr:
					ip = v.IP
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}

func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
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

// FUNCIONES DEL SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}

func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	return new(pb.Respuesta), nil
}

func (s *Server) Create(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	log.Printf("Creando registro")
	return new(pb.RespuestaAdmin), nil
}

func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	return new(pb.RespuestaAdmin), nil
}

func (s *Server) Update(ctx context.Context, message *pb.ConsultaUpdate) (*pb.RespuestaAdmin, error){
	return new(pb.RespuestaAdmin), nil
}


func main() {
	log.Printf("= INICIANDO DNS SERVER =")

	// Cargar archivo de configuración
	cargarConfig("config.json")

	// Definir e inicializar variables
	log.Printf("Inicializando variables")
	var dominioRegistro map[string]RegistroZF // relaciona el nombre de dominio con su Registro ZF respectivo
	dominioRegistro = make(map[string]RegistroZF)
	log.Printf("RegistrosZF:  %v", dominioRegistro)

	// Iniciar variables que mantenga las conexiones establecidas entre nodos
	conexionesNodos := make(map[string]*grpc.ClientConn)
	conexionesGRPC := make(map[string]pb.ServicioNodoClient)


	// Identificar el servidor DNS correspondiente a la IP de la máquina
	machineIPs := obtenerListaIPs() // Obtener lista de IPs asociadas a la máquina
	for _, dns := range config.DNS{ // Iterar sobre las IP configuradas para servidores DNS
		_, found := Find(machineIPs, dns.Ip)
		if found { // En caso de que la IP configurada coincida con alguna de las IPs de la máquina
			id := dns.Id
			ip := dns.Ip
			port := dns.Port
			conn, err := conectarNodo(ip, port)
			if err != nil{
				// Falla la conexión gRPC 
				log.Fatalf("Error al intentar realizar conexión gRPC: %s", err)
			} else {
				// Registrar servicio gRPC
				c := pb.NewServicioNodoClient(conn)
				estado, err := c.ObtenerEstado(context.Background(), new(pb.Vacio))
				if err != nil {
					//log.Fatalf("Error al llamar a ObtenerEstado(): %s", err)
					log.Printf("Nodo DNS disponible: " + id)
					iniciarNodo(port)
					break
				}
				if estado.Estado == "OK" {
					log.Printf("Almacenando conexión a nodo DNS: " + id)
					conexionesNodos[id] = conn
					conexionesGRPC[id] = c
				}
			}
		}
	}

}