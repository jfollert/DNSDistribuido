package registro

import (
	"strings"
	"log"
)

type RegistroZF struct{
	ruta string  // ruta dentro del sistema donde se almacena el archivo de Registro ZF
	rutaLog string // ruta dentro del sistema donde se almacena el archivo de Logs de Cambios.
	reloj []int32 // reloj de vector asociado al registro
	dominioLinea map[string]int // relaciona el nombre de dominio a la linea que ocupa dentro del archivo de registro
	cantLineas int
}

type Registros struct{
	dominioRegistro map[string]*RegistroZF
}



func SepararNombreDominio(nombreDominio string) (string, string) {
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