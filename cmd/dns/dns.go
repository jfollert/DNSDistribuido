package main

import (
	"fmt"
	"os"
	"log"
	"net"
	"context"
	"strings"
	"strconv"
	"errors"
	"bufio"
	"time"
	"math"
	"io"
	
	pb "github.com/jfomu/DNSDistribuido/internal/proto"
	"github.com/jfomu/DNSDistribuido/internal/config"
	"github.com/jfomu/DNSDistribuido/internal/nodo"
	//"github.com/jfomu/DNSDistribuido/internal/registros"
	"google.golang.org/grpc"
)

//// ESTRUCTURAS
type Server struct{
	nodo.Server
}

type RegistroZF struct{
	ruta string  // ruta dentro del sistema donde se almacena el archivo de Registro ZF
	rutaLog string // ruta dentro del sistema donde se almacena el archivo de Logs de Cambios.
	reloj []int32
	dominioLinea map[string]int // relaciona el nombre de dominio a la linea que ocupa dentro del archivo de registro
	cantLineas int
}


const ( //// CONSTANTES
	RUTA_REGISTROS = "registros/"
	RUTA_LOGS = "logs/"
	CONFIG_FILENAME = "config.json"
)

var ( //// VARIABLES GLOBALES
	configuracion = config.GenConfig(CONFIG_FILENAME)
	dominioRegistro map[string]*RegistroZF // relaciona el nombre de dominio con su Registro ZF respectivo
	//reg registros.Registros
	conexionesNodos map[string]*grpc.ClientConn
	conexionesGRPC map[string]pb.ServicioNodoClient
	ticker *time.Ticker
	ID_DNS string
	IP_DNS string
	PORT_DNS string
)

//// FUNCIONES
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

func separarNombreDominio(nombreDominio string) (string, string, error) {
	split := strings.Split(nombreDominio, ".")
	if len(split) == 2{
		return split[0], split[1], nil
	} 
	return "", "", errors.New(nombreDominio + " no cumple el formato, debe contener solo un punto") 
}

//// FUNCIONES DEL OBJETO SERVER
func (s *Server) ObtenerEstado(ctx context.Context, message *pb.Consulta) (*pb.Estado, error){

	if message.NombreDominio != "" && message.Ip != "" && message.Port != "" { 
		conn, err := nodo.ConectarNodo(message.Ip, message.Port)
		if err != nil{
			// Falla la conexión gRPC 
			log.Printf("Error al intentar realizar conexión gRPC: %s", err)
			return nil, err
		} 
		// Registrar servicio gRPC
		c := pb.NewServicioNodoClient(conn)
		conexionesNodos[message.NombreDominio] = conn
		conexionesGRPC[message.NombreDominio] = c
	}

	return &pb.Estado{Estado: "OK"}, nil
}

// Comando GET
func (s *Server) Get(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	// Separar nombre y el dominio en diferentes strings
	nombre, dominio, err := separarNombreDominio(message.NombreDominio)
	if err != nil{
		return nil, err
	}

	// Remover linea de registro ZF
	if registro, ok := dominioRegistro[dominio]; ok { // Verificar si se encuentra el dominio en nuestro registro ZF
		if _, ok := registro.dominioLinea[nombre]; ok { // Verificar si se encuentra la linea donde está el nombre
			
			// Abrir el archivo de registro ZF para leer y almacenar en memoria las lineas
			var readFile, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			
			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)
			
			var fileTextLines []string
			for fileScanner.Scan() {
				fileTextLines = append(fileTextLines, fileScanner.Text())
			}
			readFile.Close() // Cerramos el archivo

			linea := fileTextLines[registro.dominioLinea[nombre] - 1]

			// Verificar que la linea a leer no se encuentre vacía
			if linea == "" {
				log.Println("[ERROR] La linea del registro ZF asociada al nombre " + nombre + " está vacía")
				return nil, errors.New("La linea del registro ZF asociada al nombre " + nombre + " está vacía")
			}

			// Verificar contenido dentro de la linea a actualizar
			lineaDividida := strings.Split(linea, " IN A ")
			if len(lineaDividida) != 2 || lineaDividida[0] == "" || lineaDividida[1] == ""{
				log.Println("[ERROR] Datos corruptos en el registro ZF: " + linea)
				return nil, errors.New("Datos corruptos en el registro ZF: " + linea)
			}

			// Generamos y retornamos la respuesta a la consulta
			respuesta := new(pb.Respuesta)
			respuesta.Respuesta = lineaDividida[1]
			respuesta.Ip = IP_DNS
			respuesta.Port = PORT_DNS
			respuesta.Reloj = registro.reloj
			return respuesta, nil

		
		} else{ // Si no se encuentra la linea donde se encuentra el nombre dentro del registro ZF
			log.Printf("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
			return nil, errors.New("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
		}
	} else { //Si no se encuentra el dominio registrado
		log.Printf("No se encuentra el dominio registrado: " + dominio)
		return nil, errors.New("No se encuentra el dominio registrado: " + dominio)
	}
}

// Comando CREATE
func (s *Server) Create(ctx context.Context, message *pb.Consulta) (*pb.Respuesta, error){
	
	// Separar nombre y el dominio en diferentes strings
	nombre, dominio, err := separarNombreDominio(message.NombreDominio)
	if err != nil{
		return nil, err
	}

	salto := "\n"

	// Agregar información a registro ZF
	if _, ok := dominioRegistro[dominio]; !ok {  // Si no existe un registro ZF asociado al dominio

		// Verificar que existan las carpetas donde se almacenan los archivos de registro y log
		_, err1 := os.Stat(RUTA_REGISTROS)
		_, err2 := os.Stat(RUTA_LOGS)
		if os.IsNotExist(err1) || os.IsNotExist(err2) { 
			log.Printf("Creando directorios: %s | %s", RUTA_REGISTROS, RUTA_LOGS)
			if err = os.Mkdir(RUTA_REGISTROS, 0777); err != nil {
				return nil, err
			}
			//os.ModeDir
			if err =os.Mkdir(RUTA_LOGS, 0777); err != nil {
				return nil, err
			}
		} 

		rutaRegistros :=  RUTA_REGISTROS + ID_DNS + "/"
		rutaLogs := RUTA_LOGS + ID_DNS + "/"

		// Verificar que existan los directorios asociados al nodo
		_, err1 = os.Stat(rutaRegistros)
		_, err2 = os.Stat(rutaLogs)
		if os.IsNotExist(err1) || os.IsNotExist(err2) { // Si alguno de los archivos ya existe
			log.Printf("Creando directorios: %s | %s", rutaRegistros, rutaLogs)
			if err = os.Mkdir(rutaRegistros, 0777); err != nil {
				return nil, err
			}
			if err =os.Mkdir(rutaLogs, 0777); err != nil {
				return nil, err
			}
		}

		// Verificar que no existan los archivos asociados al registro
		_, err1 = os.Stat(rutaRegistros + dominio)
	 	_, err2 = os.Stat(rutaLogs + dominio + ".log")
		if !os.IsNotExist(err1) || !os.IsNotExist(err2) { // Si alguno de los archivos ya existe
			log.Println("Se han encotrado los archivos asociados al registro pero el registro no se encuentra en memoria.")
			return nil, errors.New("Se han encotrado los archivos asociados al registro pero el registro no se encuentra en memoria.")
		} 

		// Iniciar nuevo registro ZF en memoria
		dominioRegistro[dominio] = new(RegistroZF)
		
		// Asociar las rutas correspondientes al registro ZF
		dominioRegistro[dominio].ruta = rutaRegistros + dominio
		dominioRegistro[dominio].rutaLog = rutaLogs + dominio + ".log"

		// Inicializar variables del registro ZF
		dominioRegistro[dominio].reloj = []int32{0, 0, 0}
		dominioRegistro[dominio].dominioLinea = make(map[string]int)
		dominioRegistro[dominio].cantLineas = 0

		salto = ""

		log.Println("Se ha inicializado un nuevo registro ZF en memoria")
	}

	// Verificar que la linea del registro no exista
	if _, ok := dominioRegistro[dominio].dominioLinea[nombre]; ok {  // Si no existe una linea en el registro ZF asociada al nombre
		log.Println("Se ha intentando registrar un nombre de dominio ya existente")
		return nil, errors.New("El registro que se intenta agregar ya existe en este servidor")
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

	//Actualizar reloj de vector
	id, err := strconv.Atoi(string(ID_DNS[3]))
	if err != nil {
		log.Println(err)
		return nil, err
	}
	dominioRegistro[dominio].reloj[id - 1] += 1
	
	// Actualizar map de nombre a la linea en que se encuentra
	dominioRegistro[dominio].dominioLinea[nombre] = dominioRegistro[dominio].cantLineas

	//Generar respuesta y retornarla
	respuesta := new(pb.Respuesta) 
	respuesta.Reloj = dominioRegistro[dominio].reloj
	respuesta.Ip = IP_DNS
	respuesta.Port = PORT_DNS

	return respuesta, nil
}

// Comando DELETE
func (s *Server) Delete(ctx context.Context, message *pb.ConsultaAdmin) (*pb.RespuestaAdmin, error){
	// Separar nombre y el dominio en diferentes strings
	nombre, dominio, err := separarNombreDominio(message.NombreDominio)
	if err != nil{
		return nil, err
	}

	// Remover linea de registro ZF
	if registro, ok := dominioRegistro[dominio]; ok { // Verificar si se encuentra el dominio en nuestro registro ZF
		if _, ok := registro.dominioLinea[nombre]; ok { // Verificar si se encuentra la linea donde está el nombre
			
			// Abrir el archivo de registro ZF para leer y almacenar en memoria las lineas
			var readFile, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			
			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)
			
			var fileTextLines []string
			for fileScanner.Scan() {
				fileTextLines = append(fileTextLines, fileScanner.Text())
			}
		
			readFile.Close() // Cerramos el archivo

			// Verificar que la linea a borrar no se encuentre vacía
			lineaBorrar := dominioRegistro[dominio].dominioLinea[nombre] - 1
			if fileTextLines[lineaBorrar] == "" {
				log.Println("[ERROR] La linea del registro ZF asociada al nombre " + nombre + " ya está vacía")
				return nil, errors.New("La linea del registro ZF asociada al nombre " + nombre + " ya está vacía")
			}

			// Verificar consistencia del tamaño de las lineas leidas y las lineas del registro zf
			diferencia := dominioRegistro[dominio].cantLineas - len(fileTextLines)
			if diferencia != 0 {
				for i := 0; i < diferencia; i++ {
					fileTextLines = append(fileTextLines, "")
				}
			}

			// Crear un nuevo archivo en blanco para el registro ZF
			file1, err := os.Create(dominioRegistro[dominio].ruta)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer file1.Close()

			fileTextLines[lineaBorrar] = ""

			_, err = file1.WriteString(strings.Join(fileTextLines, "\n"))
			if err != nil {
				log.Println(err)
				return nil, err
			}

		
		} else{ // Si no se encuentra la linea donde se encuentra el nombre dentro del registro ZF
			log.Printf("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
			return nil, errors.New("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
		}
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
	nombre, dominio, err := separarNombreDominio(message.NombreDominio)
	if err != nil{
		return nil, err
	}

	// Actualizar linea de registro ZF
	if registro, ok := dominioRegistro[dominio]; ok { // Verificar si se encuentra el dominio en nuestro registro ZF
		if _, ok := registro.dominioLinea[nombre]; ok { // Verificar si se encuentra la linea donde está el nombre
			
			// Abrir el archivo de registro ZF para leer y almacenar en memoria las lineas
			var readFile, err = os.OpenFile(dominioRegistro[dominio].ruta, os.O_RDWR, 0644)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			
			fileScanner := bufio.NewScanner(readFile)
			fileScanner.Split(bufio.ScanLines)
			
			var fileTextLines []string
			for fileScanner.Scan() {
				fileTextLines = append(fileTextLines, fileScanner.Text())
			}
		
			readFile.Close() // Cerramos el archivo

			// Verificar que la linea a actualizar no se encuentre vacía
			lineaActualizar := dominioRegistro[dominio].dominioLinea[nombre] - 1
			if fileTextLines[lineaActualizar] == "" {
				log.Println("[ERROR] La linea del registro ZF asociada al nombre " + nombre + " está vacía")
				return nil, errors.New("La linea del registro ZF asociada al nombre " + nombre + " está vacía")
			}

			// Verificar contenido dentro de la linea a actualizar
			lineaVieja := strings.Split(fileTextLines[lineaActualizar], " IN A ")
			if len(lineaVieja) != 2 || lineaVieja[0] == "" || lineaVieja[1] == ""{
				log.Println("[ERROR] Datos corruptos en el registro ZF: " + fileTextLines[lineaActualizar])
				return nil, errors.New("Datos corruptos en el registro ZF: " + fileTextLines[lineaActualizar])
			}

			ip := lineaVieja[1]
			nombreOriginal := nombre
			
			// Actualizar los valores requeridos
			var cambio string
			if message.Opcion == "ip" {
				ip = message.Param
				cambio = ip
			} else if message.Opcion == "name" {
				nombre = message.Param
				cambio = nombre + "." + dominio
			}
			
			// Generar la nueva linea que se insertará en el registro ZF e insertarla
			lineaNueva := fmt.Sprintf("%s.%s IN A %s", nombre, dominio, ip)
			fmt.Println(lineaNueva)
			fileTextLines[lineaActualizar] = lineaNueva

			// Crear un nuevo archivo en blanco para el registro ZF
			file1, err := os.Create(dominioRegistro[dominio].ruta)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer file1.Close()

			// Escribir en el archivo las nuevas lineas
			_, err = file1.WriteString(strings.Join(fileTextLines, "\n"))
			if err != nil {
				log.Println(err)
				return nil, err
			}

			// Agregar información a Log de cambios
			logFile, err := os.OpenFile(dominioRegistro[dominio].rutaLog,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
				return nil, err
			}
			defer logFile.Close()
			if _, err := logFile.WriteString("\n" + "update " + nombreOriginal + "." + dominio + " " + cambio); err != nil {
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
			dominioRegistro[dominio].dominioLinea[nombre] = lineaActualizar + 1
		
			// Generar respuesta y retornarla
			respuesta := new(pb.RespuestaAdmin)
			respuesta.Reloj = dominioRegistro[dominio].reloj 
			return respuesta, nil

		} else{ // Si no se encuentra la linea donde se encuentra el nombre dentro del registro ZF
			log.Printf("[ERROR] No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
			return nil, errors.New("No es posible encontrar en el registro ZF la linea del nombre: " + nombre)
		}
	} else { //Si no se encuentra el dominio registrado
		log.Printf("[ERROR] No se encuentra el dominio registrado: " + dominio)
		return nil, errors.New("No se encuentra el dominio registrado: " + dominio)
	}
}


func (s *Server) GetFile(message *pb.Consulta, srv pb.ServicioNodo_GetFileServer) error{
	// Verificar que se recibio un dominio
	if message.NombreDominio == "" {
		return errors.New("No se ha especificado el dominio en la consulta")
	}
	
	//Abrir archivo correspondiente
	file, err := os.Open(dominioRegistro[message.NombreDominio].ruta)
	if err != nil {
		log.Println(err)
		return err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	var fileSize int64 = fileInfo.Size()
	const fileChunk = 1 * (1 << 20) 
	totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
	
	log.Printf("Registro dividido en %d piezas.\n", totalPartsNum)

	resp := new(pb.File)
	resp.FileInfo = dominioRegistro[message.NombreDominio].ruta
	for i := uint64(0); i < totalPartsNum; i++ {

		partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
		partBuffer := make([]byte, partSize)

		file.Read(partBuffer)
		resp.ChunkData = partBuffer
		srv.Send(resp)	
	}
	return nil
}

func (s *Server) SetFile(stream pb.ServicioNodo_SetFileServer) error{
	return errors.New("Función SetFile() no implementada para este nodo.")
}

func (s *Server) GetDominios(ctx context.Context, message *pb.Vacio) (*pb.Dominios, error){
	ticker.Stop()
	dominios := make([]string, 0, len(dominioRegistro))
	for d := range dominioRegistro {
        dominios = append(dominios, d)
	}
	return &pb.Dominios{Dominios: dominios}, nil
}


func main() {
	log.Printf("= INICIANDO DNS SERVER =")

	// Cargar archivo de configuración
	// configuracion = config.GenConfig(CONFIG_FILENAME)

	//reg.Init(ID_DNS)

	// Inicializar variables
	log.Printf("Inicializando variables")
	dominioRegistro = make(map[string]*RegistroZF)
	ID_DNS = ""
	IP_DNS = ""
	PORT_DNS = ""


	// Iniciar variables que mantenga las conexiones establecidas entre nodos
	conexionesNodos = make(map[string]*grpc.ClientConn)
	conexionesGRPC = make(map[string]pb.ServicioNodoClient)

	// Identificar el servidor DNS correspondiente a la IP de la máquina
	machineIPs := obtenerListaIPs() // Obtener lista de IPs asociadas a la máquina
	var id string
	var ip string
	var port string
	for _, dns := range configuracion.DNS{ // Iterar sobre las IP configuradas para servidores DNS
		id = dns.Id
		ip = dns.Ip
		port = dns.Port

		// Intentar establecer conexión con el nodo DNS
		conn, err := nodo.ConectarNodo(ip, port)
		if err != nil{ // Si falla la conexión con el nodo 
			log.Printf("Error al intentar conectar al nodo %s | %s:%s | %s\n", id, ip, port, err)
			continue
		}

		c := pb.NewServicioNodoClient(conn)
		_, err = c.ObtenerEstado(context.Background(), new(pb.Consulta))
		if err != nil { // Si el servidor no responde a una consulta gRPC
			//log.Printf("Nodo %s con servidor DNS INACTIVO\n", id )

			// Verificar si el nodo corresponde al asignado
			_, found := Find(machineIPs, dns.Ip)
			if found { // Si corresponde se asignan las variables para identificar el nodo
				ID_DNS = id
				IP_DNS = ip
				PORT_DNS = port

				// Presentarse a los otros nodos
				infoNodo := &pb.Consulta{NombreDominio: ID_DNS, Ip: IP_DNS, Port: PORT_DNS}
				for _, c := range conexionesGRPC{
					c.ObtenerEstado(context.Background(), infoNodo)
				}

				go iniciarNodo(PORT_DNS)

				//log.Println("Iniciando Timer")
				ticker = time.NewTicker(5 * time.Minute)
				quit := make(chan struct{})
				
				for {
				select {
					case <- ticker.C:
						log.Println("Coordinando servidores DNS")
						ticker.Stop()
						for _, dns := range conexionesGRPC{
							// Obtener dominios registrados en el servidor dns
							respuesta, err := dns.GetDominios(context.Background(), new(pb.Vacio))
							if err != nil{
								log.Printf("Error al ejecutar GetDominios: %s\n", err)
								continue
							}
							//log.Printf("Dominios registrados: %+v", respuesta.Dominios)

							// Revisar si existen los dominios en el nodo dominante
							for _, dom := range respuesta.Dominios{
								if _, ok := dominioRegistro[dom]; !ok { // Si no existe el dominio
									// Obtener los archivos de registros correspondientes
									stream, err := dns.GetFile(context.Background(), &pb.Consulta{NombreDominio: dom})
									if err != nil {
										log.Fatalf("Error abriendo el Stream %v", err)
									}

									done := make(chan bool)

									nombreArchivo := RUTA_REGISTROS + ID_DNS + "/" + dom
									file, err := os.OpenFile(nombreArchivo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
										if err != nil {
												fmt.Println(err)
												os.Exit(1)
										}

									go func() {
										for {
											resp, err := stream.Recv()
											if err == io.EOF {
												file.Close()
												done <- true 
												return
											}
											if err != nil {
												log.Fatalf("Error al recibir %v", err)
											}

											_, err = file.Write(resp.ChunkData)
											if err != nil {
													fmt.Println(err)
													os.Exit(1)
											}
											file.Sync() 
										}
									}()

									<-done
								}
							}

						}
						
						
					case <- quit:
						ticker.Stop()
						break
					}
				}
			}
		} else { // Si el servidor responde la consulta gRPC
			//log.Printf("Nodo %s con servidor DNS ACTIVO, almacenando conexión\n", id )
			conexionesNodos[id] = conn
			conexionesGRPC[id] = c
		}
	}

}