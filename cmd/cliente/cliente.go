package main

import (
	"log"
	"context"
	"bufio"
	"fmt"
	"os"
	"strings"

	pb "github.com/jfomu/DNSDistribuido/internal/proto"
	"github.com/jfomu/DNSDistribuido/internal/config"
	"github.com/jfomu/DNSDistribuido/internal/nodo"
	//"google.golang.org/grpc"
)

//// ESTRUCTURAS
type RegistroConsulta struct {
	Reloj []int32
	IP string
	Port string
}

//// VARIABLES GLOBALES
var configuracion *config.Config
var dominioConsulta map[string]*RegistroConsulta

func main() {

	log.Printf("= INICIANDO CLIENTE =\n")

	// Cargar archivo de configuración
	configuracion = config.GenConfig("config.json")

	// Inicializar variables
	log.Printf("Inicializando variables")
	dominioConsulta = make(map[string]*RegistroConsulta)
	

	// Conectando con el Broker
	conn, err := nodo.ConectarNodo(configuracion.Broker.Ip, configuracion.Broker.Port)
	if err != nil {
		log.Fatalf(err.Error())
	}
	broker := pb.NewServicioNodoClient(conn)

	//log.Printf("Conectado al nodo " + ip + ":" + port)
	_, err := broker.ObtenerEstado(context.Background(), new(pb.Consulta))
	if err != nil {
		log.Fatalf("Error al llamar a ObtenerEstado(): %s", err)
	}
	//log.Printf("Estado del nodo Broker: " + estado.Estado)

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
			//log.Println("nombreDominio: " + consulta.NombreDominio)
			consulta.Ip = ""
			consulta.Port = ""
			resp, err := broker.Get(context.Background(), consulta)
			if err != nil {
				log.Printf("Error al llamar a Get(): %s\n", err)
				continue
			}
			
			
			// Verificar la respuesta obtenida con el registro en memoria
			if _, ok := dominioConsulta[words[0]]; !ok { // Si no existe un registro para la consulta de ese dominio
				log.Println("Registrada consulta en memoria")
				dominioConsulta[words[0]] = &RegistroConsulta{IP: resp.Ip, Port: resp.Port, Reloj: resp.Reloj}
			} else{ // Si existe el registro para la consulta de ese dominio
				//log.Println("Registro encontrado en memoria")
				// Comparar relojes
				//log.Printf("Reloj memoria: %+v", dominioConsulta[words[0]].Reloj)
				//log.Printf("Reloj consulta: %+v", resp.Reloj)
				for i, valor := range dominioConsulta[words[0]].Reloj {
					if valor > resp.Reloj[i] {
						// Realizar consistencia
						//log.Println("REALIZAR CONSISTENCIA")
						consulta.Ip = dominioConsulta[words[0]].IP
						consulta.Port = dominioConsulta[words[0]].Port
						resp, err = broker.Get(context.Background(), consulta)
						if err != nil {
							log.Printf("Error al llamar a Get(): %s\n", err)
							continue
						}
						//log.Printf("IP: %s, Reloj: %v", resp.Respuesta, resp.Reloj)
					}
				}
			}

			log.Printf("IP: %s, Reloj: %v", resp.Respuesta, resp.Reloj)
			
			
		  
		} else {
			fmt.Println("Uso:\n get <nombre>.<dominio>")
		}	
	
	  }
}