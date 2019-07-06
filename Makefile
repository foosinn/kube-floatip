all:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" ./cmd/floatip
	strip floatip
	upx floatip
	sudo docker build . -t foosinn/floatip
	sudo docker push foosinn/floatip
