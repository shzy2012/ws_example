
build:
	env GOOS=linux GOARCH=amd64 go build -mod=vendor -o ws main.go
scp:build
	scp ./ws root@39.96.21.121:/home/works/zhouyi/ws
run:
	./ws -p=8004