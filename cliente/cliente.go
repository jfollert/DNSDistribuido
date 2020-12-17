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
	c := pb.NewServicioNodoClient(conn)

	//log.Printf("Conectado al nodo " + ip + ":" + port)
	estado, err := c.ObtenerEstado(context.Background(), new(pb.Vacio))
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
		  fmt.Println("create")
		} else if strings.Compare("update", words[0]) == 0 {
			fmt.Println("update")
		} else if strings.Compare("delete", words[0]) == 0 {
			fmt.Println("delete")
		} else {
			fmt.Println("Usage:\n create\n update\n delete")
		}
	
	  }
}