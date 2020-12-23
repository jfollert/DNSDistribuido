package main

import (
	"log"
	"context"
	"bufio"
	"fmt"
	"os"
	"strings"
	"io/ioutil"
	"encoding/json"


	pb "../proto"
	"google.golang.org/grpc"
)

//// ESTRUCTURAS
type RegistroConsulta struct {
	Reloj []int32
	IP string
	Port string
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
var config Config
var dominioConsulta map[string]*RegistroConsulta

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

func conectarNodo(ip string, port string) *grpc.ClientConn {
	var conn *grpc.ClientConn
	log.Printf("Intentando iniciar conexión con " + ip + ":" + port)
	host := ip + ":" + port
	conn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}
	return conn
}


func main() {

	log.Printf("= INICIANDO CLIENTE =\n")

	// Cargar archivo de configuración
	cargarConfig("config.json")

	// Inicializar variables
	log.Printf("Inicializando variables")
	dominioConsulta = make(map[string]*RegistroConsulta)
	

	// Conectando con el Broker
	conn := conectarNodo(config.Broker.Ip, config.Broker.Port)
	broker := pb.NewServicioNodoClient(conn)

	//log.Printf("Conectado al nodo " + ip + ":" + port)
	estado, err := broker.ObtenerEstado(context.Background(), new(pb.Vacio))
	if err != nil {
		log.Fatalf("Error al llamar a ObtenerEstado(): %s", err)
	}
	log.Printf("Estado del nodo Broker: " + estado.Estado)

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		text = strings.ToLower(text)
		words := strings.Split(text, " ")


		if strings.Compare("get", words[0]) == 0 && len(words) == 2  { // Si el comando ingresado es get
			consulta := new(pb.Consulta)
			consulta.NombreDominio = words[1]
			log.Println("nombreDominio: " + consulta.NombreDominio)
			consulta.Ip = ""
			consulta.Port = ""
			resp, err := broker.Get(context.Background(), consulta)
			if err != nil {
				log.Printf("Error al llamar a Get(): %s\n", err)
				continue
			}
			log.Printf("IP: %s, Reloj: %v", resp.Respuesta, resp.Reloj)
			
			// Verificar la respuesta obtenida con el registro en memoria
			log.Println("Registrar consulta en memoria")
			if _, ok := dominioConsulta[words[0]]; !ok { // Si no existe un registro para la consulta de ese dominio
				dominioConsulta[words[0]] = &RegistroConsulta{IP: resp.Ip, Port: resp.Port, Reloj: resp.Reloj}
			} else{ // Si existe el registro para la consulta de ese dominio
				//EDITAR
				dominioConsulta[words[0]] = &RegistroConsulta{IP: resp.Ip, Port: resp.Port, Reloj: resp.Reloj}
			}
			
			
		  
		} else {
			fmt.Println("Uso:\n get <nombre>.<dominio>")
		}	
	
	  }
}