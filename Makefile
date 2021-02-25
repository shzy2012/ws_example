
build:
	env GOOS=linux GOARCH=amd64 go build -mod=vendor -o ws_example main.go
scp:build
	scp ./ws_example root@39.96.21.121:/home/works/zhouyi/ws
run:
	./ws_example -p=8004