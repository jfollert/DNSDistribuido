# Sistema DNS Distribuido con consistencia de datos

## Ejecución
1. Ejecutar los servidores DNS en sus respectivas máquinas utilizando el comando:
```console
make dns
```

2. Ejecutar el broker en su respectiva máquina utilizando el comando:
```console
make broker
```

3. Ejecutar el nodo administrador en su respectiva máquina con el comando:
```console
make admin
```

4. Ejecutar el nodo cliente en su respectiva máquina con el comando:
```console
make cliente
```

### Admin
El nodo administrador puede recibir los comandos:
- **create** *\<nombre\>.\<dominio\> \<IP\>*
- **delete** *\<nombre\>.\<dominio\>*
- **update** *\<nombre\>.\<dominio\> \<opción\> \<parámetro\>*

Los cuales se verán reflejados en los directorios *registros/* y *logs/* en los respectivos servidores DNS donde se apliquen los comandos.

### Cliente
El nodo cliente puede recibir el comando:
- **get** *\<nombre\>.\<dominio\>*

## Consideraciones
- Todos los nombres de dominios deben seguir la estructura *nombre.dominio*, una mayor cantidad de puntos causará errores.
  
- Para limpiar los archivos generados en *registros/* y *logs/* se puede utilizar el comando
```console
make clean
```


## ProtocolBuffers
En caso de necesitar recompilar los archivo *.proto* se deben seguir los siguientes pasos:

1. Se debe agregar la ruta donde se encuentra *protoc-gen-go* al PATH
```console
export PATH="$PATH:$(go env GOPATH)/bin"
```
2. Para compilar los archivos .proto
```console
protoc -I internal/proto --go_out=plugins=grpc:internal/proto internal/proto/*.proto
```