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

	log.Printf("Inicializando variables")
	
	
	log.Println("Estableciendo conexión con el Broker")
	conn := conectarNodo("127.0.0.1", "9000")
	broker := pb.NewServicioNodoClient(conn)

	//log.Printf("Conectado al nodo " + ip + ":" + port)
	estado, err := broker.ObtenerEstado(context.Background(), new(pb.Vacio))
	if err != nil {
		log.Fatalf("Error al llamar a ObtenerEstado(): %s", err)
	}
	log.Printf("Estado del nodo seleccionado: " + estado.Estado)

	//Receive Command
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		words := strings.Split(text, " ")
	
		if strings.Compare("create", words[0]) == 0 { 
			if len(words) != 3 {
				fmt.Printf("[ERROR] Usage:\n create <nombre>.<dominio> <IP>\n")
			} else {
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
					log.Fatalf("Error al llamar a Create(): %s", err)
					}
				log.Printf("Create exitoso! - Reloj: %+v", dnsResp.Reloj)
			}

		} else if strings.Compare("update", words[0]) == 0 {
			fmt.Println("update")

		} else if strings.Compare("delete", words[0]) == 0 {
			if len(words) != 2 {
				fmt.Printf("[ERROR] Usage:\n delete <nombre>.<dominio>\n")
			} else {
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
					log.Fatalf("Error al llamar a Delete(): %s", err)
					}
				log.Printf("Delete exitoso! - Reloj: %+v", dnsResp.Reloj)
			}
		} else {
			fmt.Println("Usage:\n create\n update\n delete")
		}
	
	  } 
}