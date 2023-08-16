# azure-storage
video-storage microservice

go version
go version go1.20.4 linux/amd64

go mod init github.com/michaelspinks/azure-storage
go mod tidy

make run
make build
make clean
make test

go get github.com/Azure/azure-storage-blob-go/azblob


docker build -t video-storage --file Dockerfile .
docker run -p 4000:4000 video-storage

001 - init