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
//Revisa si el reloj local mayor a la de la primera respuesta entregada por un dns al azar
func compararRelojes(local, respuesta []int32) bool{
	for i := 0; i<3; i++{
		if local[i]>respuesta[i]{
			return true
		}}
	return false
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

	conn := conectarNodo("127.0.0.1", "9000")
	broker := pb.NewServicioNodoClient(conn)

	//log.Printf("Conectado al nodo " + ip + ":" + port)
	estado, err := broker.ObtenerEstado(context.Background(), new(pb.Vacio))
	if err != nil {
		log.Fatalf("Error al llamar a ObtenerEstado(): %s", err)
	}
	log.Printf("Estado del nodo seleccionado: " + estado.Estado)
	
	//Registro de memoria del cliente
	//memoriaLocal:= make(map[string]*pb.Respuesta)
	ipLocal:=make(map[string]string)
	relojLocal:=make(map[string][]int32)
	
	//Receive Command
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		words := strings.Split(text, " ")
	
		if strings.Compare("get", words[0]) == 0  { 
			if len(words) != 2 {
				fmt.Printf("[ERROR] Usage:\n get <nombre>.<dominio> \n")
			} else {
				cons := new(pb.Consulta)
				cons.NombreDominio = words[1]
				resp, err := broker.Get(context.Background(), cons)
				if err != nil {
				log.Fatalf("Error al llamar a Get(): %s", err)
				}
				//Aca se aplica consistencia
				
				if compararRelojes(relojLocal[words[1]], resp.Reloj){
					//Se prepara la siguiente consulta
					cons := new(pb.Consulta)
					cons.NombreDominio = words[1]
					//Especificando una ip
					cons.Ip= ipLocal[words[1]]
					respu, err := broker.Get(context.Background(), cons)
					if err != nil {
					log.Fatalf("Error al llamar a Get(): %s", err)
					}
					//Se guarda en primer caso
					ipLocal[words[1]]= respu.Ip
					relojLocal[words[1]]= respu.Reloj
				}
				else{
					//O se guarda la respuesta original, en caso de que bastara.
					ipLocal[words[1]]= resp.Ip
					relojLocal[words[1]]= resp.Reloj
				}
			}
		}else {
			fmt.Println("Usage:\n get")
		}
	
	  }
}
