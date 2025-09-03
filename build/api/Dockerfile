# Usa una imagen base de Go
FROM golang:1.24-alpine

# Establece el directorio de trabajo
WORKDIR /server

# Copia los archivos del proyecto
COPY . .

# Descarga las dependencias
RUN go mod download

# Compila la aplicación
RUN go build -o server ./cmd

# Crea el directorio para el filestore
RUN mkdir -p /server/filestore

# Expone el puerto de la aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./server"]