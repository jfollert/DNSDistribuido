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
	//"io"
	"errors"

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
	cantLineas int
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

func separarNombreDominio(nombreDominio string) (string, string) {
	split := strings.Split(nombreDominio, ".")
	var nombre string
	var dominio string

	if len(split) == 2{
	nombre = split[0]
	dominio = split[1]
	} else {
		log.Fatal("[ERROR] Error dividiendo la variable NombreDominio")
	}
	return nombre, dominio
}

//// FUNCIONES DEL OBJETO SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Vacio) (*pb.Estado, error){
	estado := new(pb.Estado)
	estado.Estado = "OK"
	return estado, nil
}

// Comando GET
func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	return new(pb.Respuesta), nil
}

// Comando CREATE
func (s *Server) Create(ctx context.Context, message *pb.Consulta) (*pb.RespuestaAdmin, error){
	// Separar nombre y el dominio en diferentes strings
	nombre, dominio := separarNombreDominio(message.NombreDominio)
	salto := "\n"

	// Agregar información a registro ZF
	if _, ok := dominioRegistro[dominio]; !ok {  // Si no existe un registro ZF asociado al dominio
		rutaRegistro := "dns/registros/" + ID_DNS + "_" + dominio + ".zf"
		rutaLog := "dns/logs/" + ID_DNS + "_" + dominio + ".log"
		
		// Verificar que no existan los archivos asociados al registro
		var _, err1 = os.Stat(rutaRegistro)
		var _, err2 = os.Stat(rutaLog)
		if !os.IsNotExist(err1) || !os.IsNotExist(err2) { // Si alguno de los archivos ya existe
			log.Println("Se han encotrado los archivos asociados al registro pero el registro no se encuentra en memoria.")
			return nil, errors.New("Se han encotrado los archivos asociados al registro pero el registro no se encuentra en memoria.")
		} 

		// Iniciar nuevo registro ZF en memoria
		dominioRegistro[dominio] = new(RegistroZF)
		
		// Asociar las rutas correspondientes al registro ZF
		dominioRegistro[dominio].ruta = rutaRegistro
		dominioRegistro[dominio].rutaLog = rutaLog

		// Inicializar variables del registro ZF
		dominioRegistro[dominio].reloj = []int32{0, 0, 0}
		dominioRegistro[dominio].dominioLinea = make(map[string]int)
		dominioRegistro[dominio].cantLineas = 0
		salto = ""

		log.Println("Se ha inicializado un nuevo registro ZF en memoria")
	}

	// Agregar información a archivo de registro ZF
	regFile, err := os.OpenFile(dominioRegistro[dominio].ruta,
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer regFile.Close()
	if _, err := regFile.WriteString(salto + nombre + "." + dominio + " IN A " + message.Ip); err != nil {
		log.Println(err)
		return nil, err
	}
	dominioRegistro[dominio].cantLineas += 1
	log.Println("Información agregada al archivo del registro ZF")

	// Agregar información a Log de cambios
	logFile, err := os.OpenFile(dominioRegistro[dominio].rutaLog,
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer logFile.Close()
	if _, err := logFile.WriteString(salto + "create " + nombre + "." + dominio + " " + message.Ip); err != nil {
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
	dominioRegistro[dominio].dominioLinea[nombre] = dominioRegistro[dominio].cantLineas

	// Generar respuesta y retornarla
	respuesta := new(pb.RespuestaAdmin) 
	respuesta.Reloj = dominioRegistro[dominio].reloj 
	return respuesta, nil

}

// Comando DELETE
func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	// Separar nombre y el dominio en diferentes strings
	nombre, dominio := separarNombreDominio(message.NombreDominio)

	// Remover linea de registro ZF
	if registro, ok := dominioRegistro[dominio]; ok { // Verificar si se encuentra el dominio en nuestro registro ZF
		if _, ok := registro.dominioLinea[nombre]; ok { // Verificar si se encuentra la linea donde está el nombre
			var file, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer file.Close()

			text, err := ioutil.ReadAll(file)
			
			file1, err := os.Create(dominioRegistro[dominio].ruta)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer file1.Close()
			lineas := strings.Split(string(text), "\n")
			flag := false
			for i, linea := range lineas{
				if i == dominioRegistro[dominio].cantLineas {
					break
				}
				if i != dominioRegistro[dominio].dominioLinea[nombre] - 1 {
					if i == 0 || flag {
						_, err = file1.WriteString(linea)
						flag = false
					} else {
						_, err = file1.WriteString("\n" + linea)
					}
				} else {
					_, err = file1.WriteString("\n")
					flag = true
				}
				if err != nil {
					log.Println(err)
					return nil, err
				}
			}
		
		} else{ // Si no se encuentra la linea donde se encuentra el nombre dentro del registro ZF
			log.Printf("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
			return nil, errors.New("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
		}
		log.Println("Linea eliminada del registro ZF")
	} else { //Si no se encuentra el dominio registrado
		log.Printf("No se encuentra el dominio registrado: " + dominio)
		return nil, errors.New("No se encuentra el dominio registrado: " + dominio)
	}
		

	// Agregar información a Log de cambios
	logFile, err := os.OpenFile(dominioRegistro[dominio].rutaLog,
	os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer logFile.Close()
	if _, err := logFile.WriteString("\n" + "delete " + nombre + "." + dominio); err != nil {
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

	// Remover mapeo de nombre a la linea en que se encuentra
	delete(dominioRegistro[dominio].dominioLinea, nombre)

	// Generar respuesta y retornarla
	respuesta := new(pb.RespuestaAdmin)
	respuesta.Reloj = dominioRegistro[dominio].reloj 
	return respuesta, nil
}

// Comando UPDATE
func (s *Server) Update(ctx context.Context, message *pb.ConsultaUpdate) (*pb.RespuestaAdmin, error){
	// Separar nombre y el dominio en diferentes strings
	nombre, dominio := separarNombreDominio(message.NombreDominio)

	// Actualizar linea de registro ZF
	if registro, ok := dominioRegistro[dominio]; ok { // Verificar si se encuentra el dominio en nuestro registro ZF
		if _, ok := registro.dominioLinea[nombre]; ok { // Verificar si se encuentra la linea donde está el nombre
			var file, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer file.Close()

			text, err := ioutil.ReadAll(file)
			
			file1, err := os.Create(dominioRegistro[dominio].ruta)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer file1.Close()
			lineas := strings.Split(string(text), "\n")
			flag := false
			for i, linea := range lineas{
				if i == dominioRegistro[dominio].cantLineas {
					break
				}
				if i != dominioRegistro[dominio].dominioLinea[nombre] - 1 {
					if i == 0 || flag {
						_, err = file1.WriteString(linea)
						flag = false
					} else {
						_, err = file1.WriteString("\n" + linea)
					}
				} else {
					_, err = file1.WriteString("\n")
					flag = true
				}
				if err != nil {
					log.Println(err)
					return nil, err
				}
			}
		
		} else{ // Si no se encuentra la linea donde se encuentra el nombre dentro del registro ZF
			log.Printf("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
			return nil, errors.New("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
		}
		log.Println("Linea eliminada del registro ZF")
	} else { //Si no se encuentra el dominio registrado
		log.Printf("No se encuentra el dominio registrado: " + dominio)
		return nil, errors.New("No se encuentra el dominio registrado: " + dominio)
	}


	// Generar respuesta y retornarla
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