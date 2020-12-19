# Sistema DNS Distribuido con consistencia de datos
- **Ramo:** Sistemas Distribuidos
- **Desarrolladores:**
  - José Follert 
  - Harold Melo

## Ejecución
1. Ejecutar los servidores DNS en sus respectivas máquinas utilizando el comando:
```console
make DNS
```

2. Ejecutar el broker en su respectiva máquina utilizando el comando:
```console
make Broker
```

3. Ejecutar el nodo administrador en su respectiva máquina con el comando:
```console
make Admin
```

4. Ejecutar el nodo cliente en su respectiva máquina con el comando:
```console
make Cliente
```

### PROTOC
Se debe agregar la ruta donde se encuentra *protoc-gen-go* al PATH
```console
export PATH="$PATH:$(go env GOPATH)/bin"
```
Para compilar los archivos .proto
```console
protoc -I proto --go_out=plugins=grpc:proto proto/*.proto
```

### CONSIDERACIONES
- Todos los nombres de dominios deben seguir la estructura *nombre.dominio*, una mayor cantidad de puntos causará errores.