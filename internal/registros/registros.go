package registros

import (
	"errors"
	"os"
	"bufio"
	"strings"
	"log"
	"strconv"
)

const ( 
	RUTA_REGISTROS = "registros/"
	RUTA_LOGS = "logs/"
)

var (
	rutaRegistros string
	rutaLogs string
	ID_DNS string
)

type RegistroZF struct{
	idDNS string
	rutaReg string  // ruta dentro del sistema donde se almacena el archivo de Registro ZF
	rutaLog string // ruta dentro del sistema donde se almacena el archivo de Logs de cambios.
	reloj []int32
	lineas map[string]int // relaciona el nombre de dominio a la linea que ocupa dentro del archivo de registro
	cantLineas int
}

type Registros map[string]*RegistroZF


func CrearDirectorio(dir string)  error {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.Mkdir(dir, 0777); err != nil {
			return err
		}
	}
	return nil
}

func SepararNombreDominio(nombreDominio string) (string, string, error) {
	split := strings.Split(nombreDominio, ".")
	if len(split) == 2{
		return split[0], split[1], nil
	} 
	return "", "", errors.New(nombreDominio + " no cumple el formato, debe contener solo un punto") 
}

func (r *RegistroZF) GetReloj() []int32 {
	return r.reloj
}

func (r *RegistroZF) ExisteNombre(nombre string) bool {
	if _, ok := r.lineas[nombre]; ok {
		return true
	}
	return false
}

func (r *RegistroZF) CargarArchivoRegistro() error {
	// Abrir el archivo de registro ZF para cargarlo en memoria
	var file, err = os.OpenFile(r.rutaReg, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	var lineasArchivo []string
	for fileScanner.Scan() {
		lineasArchivo = append(lineasArchivo, fileScanner.Text())
	}
	file.Close() // Cerramos el archivo

	if len(lineasArchivo) != 0 { // Si el archivo no está vacio
		for i, linea := range lineasArchivo {
			splitLinea := strings.Split(linea, " IN A ")
			nombre, _, err := SepararNombreDominio(splitLinea[0])
			if err != nil {
				return err
			}
			r.lineas[nombre] = i // Asociamos el nombre de dominio a la linea del archivo
			r.cantLineas += 1 // Aumentamos el contador de lineas
		}
	} else{
		log.Println("EL ARCHIVO ESTA VACIO")
	}
	return nil
}

func (r *RegistroZF) EscribirLineaLog(linea string) error {
	// Abrir archivo de logs
	file, err := os.OpenFile(r.rutaLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	// Escribir linea en el archivo de logs
	if _, err = file.WriteString(linea); err != nil {
		return err
	}
	file.Close()
	return nil
}

func (r *RegistroZF) EscribirLineaRegistro(linea string) error {
	// Abrir archivo de logs
	file, err := os.OpenFile(r.rutaReg, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	// Escribir linea en el archivo de logs
	if _, err = file.WriteString(linea); err != nil {
		return err
	}
	file.Close()
	return nil
}

func (r *RegistroZF) AvanzarReloj() error {
	idIndex, err := strconv.Atoi(string(ID_DNS))
	if err != nil {
		return err
	}
	r.reloj[idIndex - 1] += 1
	return nil
}

func (r Registros) AgregarRegistro(nombre string, dominio string, ip string) error {
	if r.ExisteRegistroMemoria(dominio) {
		if r.ExisteArchivoRegistro(dominio) {
			if !r[dominio].ExisteNombre(nombre) {
				// Si es la primera linea del registro no se agrega el salto de linea al comienzo
				var salto string
				if r[dominio].cantLineas != 0 {
					salto = ""
				} else {
					salto = "\n"
				}

				// Escribir linea en el archivo de registro
				linea := salto + nombre + "." + dominio + " IN A " + ip
				if err := r[dominio].EscribirLineaRegistro(linea); err != nil {
					return err
				}
				
				r[dominio].lineas[nombre] = r[dominio].cantLineas // Se almacena en memoria el indice de la linea donde se encuentra el nombre
				r[dominio].cantLineas += 1 // Se incrementa la cantidad de lineas
				
				// Escribir linea en el archivo de logs
				linea = salto + "create " + nombre + "." + dominio + " " + ip
				if err := r[dominio].EscribirLineaLog(linea); err != nil {
					return err
				}

				// Actualizar reloj de vector
				r[dominio].AvanzarReloj()

				return nil
				
			}
			return errors.New("El nombre que se quiere agregar ya existe en el registro")
		}
		return errors.New("No es posible encontrar el archivo de registro asociado al dominio")
	}
	return errors.New("El dominio no se encuentra registrado en memoria")
}

func (r Registros) Init(id string) error {
	rutaRegistros = RUTA_REGISTROS + id + "/"
	rutaLogs = RUTA_LOGS + id + "/"
	ID_DNS = id

	r = make(Registros)

	// Verificar que existan los directorios asociados al registro
	if err := CrearDirectorio(RUTA_REGISTROS); err != nil {
		return err
	}
	if err := CrearDirectorio(RUTA_LOGS); err != nil {
		return err
	} 
	if err := CrearDirectorio(rutaRegistros); err != nil {
		return err
	}
	if err := CrearDirectorio(rutaLogs); err != nil {
		return err
	}
	return nil
}

func (r Registros) ExisteRegistroMemoria(dominio string) bool {
    if _, ok := r[dominio]; ok {
		return true
	}
	return false
}

func (r Registros) ExisteArchivoRegistro(dominio string) bool {
	if 	_, err := os.Stat(rutaRegistros + dominio); os.IsNotExist(err){
		return false
	}
	return true
}

func (r Registros) CrearRegistro(dominio string) error {
	if !r.ExisteRegistroMemoria(dominio){
		// Iniciar nuevo registro ZF en memoria
		r[dominio] = new(RegistroZF)
			
		// Asociar las rutas correspondientes al registro ZF
		r[dominio].rutaReg = rutaRegistros + dominio
		r[dominio].rutaLog = rutaLogs + dominio + ".log"

		// Inicializar variables del registro ZF
		r[dominio].reloj = []int32{0, 0, 0}
		r[dominio].lineas = make(map[string]int)
		r[dominio].cantLineas = 0

		// Verificar que no exista el archivo de registro asociado a el dominio
		if 	!r.ExisteArchivoRegistro(dominio) { // Si no existe el archivo de registros
			// Generar archivo de registros
			regFile, err := os.Create(r[dominio].rutaReg)
			if err != nil {
				return err
			}
			regFile.Close()

			// Generar archivo de logs
			logFile, err := os.Create(r[dominio].rutaLog)
			if err != nil {
				return err
			}
			logFile.Close()
		} else { // Si existe el archivo de registro ZF
			r[dominio].CargarArchivoRegistro()
		}

		
		return nil
	}
	return errors.New("El registro del dominio " + dominio + " ya existe")
}