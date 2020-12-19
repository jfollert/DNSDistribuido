package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
	"log"
	"net"
	"context"
	"strings"
	"strconv"
	"io"

	pb "../proto"
	"google.golang.org/grpc"
)

//// ESTRUCTURAS
type Server struct{}

type RegistroZF struct{
	ruta string  // ruta dentro del sistema donde se almacena el archivo de Registro ZF
	rutaLog string // ruta dentro del sistema donde se almacena el archivo de Logs de Cambios.
	reloj []int32
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

//// VARIABLES GLOBALES
var dominioRegistro map[string]*RegistroZF // relaciona el nombre de dominio con su Registro ZF respectivo
var config Config
var ID_DNS string

//// FUNCIONES
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

func iniciarRegistroZF(dominio string) {
	dominioRegistro[dominio] = new(RegistroZF)
	log.Printf("Iniciado nuevo Registro ZF para el dominio %s", dominio)

	dominioRegistro[dominio].ruta = "dns/registros/" + ID_DNS + "_" + dominio + ".zf"
	// Verificar que no exista el archivo
	var _, err = os.Stat(dominioRegistro[dominio].ruta)
	if os.IsNotExist(err) {
		f, err := os.Create(dominioRegistro[dominio].ruta)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
	}
	log.Printf("Generado archivo Registro ZF: %s", dominioRegistro[dominio].ruta)

	dominioRegistro[dominio].rutaLog = "dns/logs/" + ID_DNS + "_" + dominio + ".log"
	// Verificar que no exista el archivo
	var _, err2 = os.Stat(dominioRegistro[dominio].rutaLog)
	if os.IsNotExist(err2) {
		f, err := os.Create(dominioRegistro[dominio].rutaLog)
		if err != nil {
			log.Println(err)
		}
		defer f.Close()
	}
	log.Printf("Generado Log de cambios: %s", dominioRegistro[dominio].rutaLog)
	
	dominioRegistro[dominio].reloj = []int32{0, 0, 0}
	dominioRegistro[dominio].dominioLinea = make(map[string]int)
	log.Printf("RegistroZF registrado de forma exitosa")

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

//// FUNCIONES DEL OBJETO SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}

func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	return new(pb.Respuesta), nil
}

func (s *Server) Create(ctx context.Context, message *pb.Consulta) (*pb.RespuestaAdmin, error){
	// Separar nombre y el dominio en diferentes strings
	split := strings.Split(message.NombreDominio, ".")
	var nombre string
	var dominio string

	if len(split) == 2{
	nombre = split[0]
	dominio = split[1]
	} else {
		log.Println("[ERROR] Error dividiendo la variable NombreDominio")
		return nil, nil
	}

	// Agregar información a registro ZF
	_, ok := dominioRegistro[dominio]
	if !ok {  // Si no existe un registro ZF asociado
		iniciarRegistroZF(dominio)
	}

	reg, err := os.OpenFile(dominioRegistro[dominio].ruta,
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer reg.Close()
	if _, err := reg.WriteString(nombre + "." + dominio + " IN A " + message.Ip + "\n"); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Información agregada al registro ZF")

	// Agregar información a Log de cambios
	logFile, err := os.OpenFile(dominioRegistro[dominio].rutaLog,
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer logFile.Close()
	if _, err := logFile.WriteString("create " + nombre + "." + dominio + " " + message.Ip + "\n"); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Información agregada al Log de cambios")

	// Actualizar reloj de vector
	id, err := strconv.Atoi(string(ID_DNS[3]))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	dominioRegistro[dominio].reloj[id - 1] += 1
	
	// Actualizar map de nombre a la linea en que se encuentra
	dominioRegistro[dominio].dominioLinea[nombre] = len(dominioRegistro[dominio].dominioLinea) + 1

	respuesta := new(pb.RespuestaAdmin) 
	respuesta.Reloj = dominioRegistro[dominio].reloj 
	return respuesta, nil
}

func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	// Separar nombre y el dominio en diferentes strings
	split := strings.Split(message.NombreDominio, ".")
	var nombre string
	var dominio string

	if len(split) == 2{
	nombre = split[0]
	dominio = split[1]
	} else {
		log.Println("[ERROR] Error dividiendo la variable NombreDominio")
		return nil, nil
	}

	// Remover linea de registro ZF
	_, ok := dominioRegistro[dominio].dominioLinea[nombre]
	if ok {  // Si se encuentra la linea donde está el nombre
		var file, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer file.Close()

		// Read file, line by line
		var text = make([]byte, 1024)
		for {
			_, err = file.Read(text)
			if err == io.EOF {
				break
			}
			if err != nil && err != io.EOF {
				log.Println(err)
				break
			}
		}
		
		file, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer file.Close()
		lineas := strings.Split(strings.TrimSpace(string(text)), "\n")
    	for i, linea := range lineas{
			if i != dominioRegistro[dominio].dominioLinea[nombre] - 1 {
				_, err = file.WriteString(linea+"\n")
				if err != nil {
					log.Println(err)
					return nil, err
				}
			} else {
				_, err = file.WriteString("\n")
				if err != nil {
					log.Println(err)
					return nil, err
				}
			}	
		}

	} else{
		log.Printf("No se encuentra registrada la linea")
	}
	log.Println("Linea eliminada del registro ZF")

	// Agregar información a Log de cambios
	logFile, err := os.OpenFile(dominioRegistro[dominio].rutaLog,
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer logFile.Close()
	if _, err := logFile.WriteString("delete " + nombre + "." + dominio + "\n"); err != nil {
		log.Println(err)
		return nil, err
	}
	log.Println("Información agregada al Log de cambios")


	// Actualizar reloj de vector
	id, err := strconv.Atoi(string(ID_DNS[3]))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	dominioRegistro[dominio].reloj[id - 1] += 1
	log.Println("Reloj actualizado")

	respuesta := new(pb.RespuestaAdmin)
	respuesta.Reloj = dominioRegistro[dominio].reloj 
	return respuesta, nil
}

func (s *Server) Update(ctx context.Context, message *pb.ConsultaUpdate) (*pb.RespuestaAdmin, error){
	respuesta := new(pb.RespuestaAdmin)
	return respuesta, nil
}


func main() {
	log.Printf("= INICIANDO DNS SERVER =")

	// Cargar archivo de configuración
	cargarConfig("config.json")

	// Definir e inicializar variables
	log.Printf("Inicializando variables")
	dominioRegistro = make(map[string]*RegistroZF)
	ID_DNS = ""

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
					ID_DNS = id
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