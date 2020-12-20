package main

import (
	"log"
	"context"
	"bufio"
	"fmt"
	"os"
	"strings"

	pb "../proto"
	"google.golang.org/grpc"
)

// ESTRUCTURAS
type Cambio struct {
	cambio string
	reloj [3]int
	ipDNS string
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
	log.Printf("= INICIANDO ADMIN =\n")

	//log.Printf("Inicializando variables")
	
	log.Println("Estableciendo conexión con el Broker")
	conn := conectarNodo("127.0.0.1", "9000")
	broker := pb.NewServicioNodoClient(conn)

	//log.Printf("Conectado al nodo " + ip + ":" + port)
	estado, err := broker.ObtenerEstado(context.Background(), new(pb.Vacio))
	if err != nil {
		log.Fatalf("Error al llamar a ObtenerEstado(): %s", err)
	}
	log.Printf("Estado del nodo seleccionado: " + estado.Estado)

	// Recibir comando por la terminal
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		words := strings.Split(text, " ")
		
		//// Comando CREATE
		if strings.Compare("create", words[0]) == 0{ 
			// Verificar el número de parámetros y puntos en el nombre de dominio
			if len(words) != 3 ||  len(strings.Split(words[1], ".")) != 2 {
				log.Printf("[ERROR] Usar:\n\t create <nombre>.<dominio> <IP>\n")
				continue
			}

			// Solicitar un servidor DNS aleatorio al Broker
			resp, err := broker.Get(context.Background(), new(pb.Consulta))
			if err != nil {
			log.Fatalf("Error al llamar a Get(): %s", err)
			}

			log.Println("Estableciendo conexión con el nodo DNS")	
			conn := conectarNodo(resp.Ip, resp.Port)
			dns := pb.NewServicioNodoClient(conn)

			consulta := new(pb.Consulta)
			consulta.NombreDominio = words[1]
			consulta.Ip = words[2]

			dnsResp, err := dns.Create(context.Background(), consulta)
			if err != nil {
				log.Printf("Error al llamar a Create(): %s", err)
				continue
				}
			log.Printf("Create exitoso! - Reloj: %+v", dnsResp.Reloj)


		//// Comando UPDATE
		} else if strings.Compare("update", words[0]) == 0 {
			// Verificar el número de parámetros y puntos en el nombre de dominio
			if len(words) != 4 ||  len(strings.Split(words[1], ".")) != 2  { 
				log.Printf("[ERROR] Usar:\n\t update <nombre>.<dominio> <opcion> <parámetro>\n\t <opcion> puede tomar los valores de ip o name\n")
				continue
			}

			// Solicitar un servidor DNS aleatorio al Broker
			resp, err := broker.Get(context.Background(), new(pb.Consulta))
			if err != nil {
			log.Fatalf("Error al llamar a Get(): %s", err)
			}

			log.Println("Estableciendo conexión con el nodo DNS")
			conn := conectarNodo(resp.Ip, resp.Port)
			dns := pb.NewServicioNodoClient(conn)

			consulta := new(pb.ConsultaUpdate)
			consulta.NombreDominio = words[1]
			consulta.Opcion = words[2]
			consulta.Param = words[3]

			dnsResp, err := dns.Update(context.Background(), consulta)
			if err != nil {
				log.Printf("Error al llamar a Update(): %s", err)
				continue
				}
			log.Printf("Update exitoso! - Reloj: %+v", dnsResp.Reloj)
			

		//// Comando DELETE
		} else if strings.Compare("delete", words[0]) == 0 {
			// Verificar el número de parámetros y puntos en el nombre de dominio
			if len(words) != 2 ||  len(strings.Split(words[1], ".")) != 2 {
				log.Printf("[ERROR] Usar:\n\t delete <nombre>.<dominio>\n")
				continue
			}
			resp, err := broker.Get(context.Background(), new(pb.Consulta))
			if err != nil {
			log.Fatalf("Error al llamar a Get(): %s", err)
			}

			log.Println("Estableciendo conexión con el nodo DNS")
			conn := conectarNodo(resp.Ip, resp.Port)
			dns := pb.NewServicioNodoClient(conn)

			consulta := new(pb.ConsultaAdmin)
			consulta.NombreDominio = words[1]

			dnsResp, err := dns.Delete(context.Background(), consulta)
			if err != nil {
				log.Printf("Error al llamar a Delete(): %s", err)
				continue
				}
			log.Printf("Delete exitoso! - Reloj: %+v", dnsResp.Reloj)
			
		} else { // En caso de no recibir un comando válido
			fmt.Println("Usar:\n\t create <nombre>.<dominio> <IP>\n\t update <nombre>.<dominio> <opción> <parámetro>\n\t delete <nombre>.<dominio>")
		}
	
	  } 
}