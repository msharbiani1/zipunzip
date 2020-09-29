all: zipunzip zipunzip-linux

zipunzip: zipunzip.go
	go build -o zipunzip
zipunzip-linux: zipunzip.go
	GOOS=linux go build -o zipunzip-linux
docker-image:
	docker build .
