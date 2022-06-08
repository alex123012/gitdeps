.PHONY: build install clean

build:
	CGO_ENABLED=0 go build -o bin/dependency-bot github.com/alex123012/dependency-bot/cmd/dependency-bot
build-race:
	CGO_ENABLED=1 go build -race -o bin/dependency-bot github.com/alex123012/dependency-bot/cmd/dependency-bot
install:
	CGO_ENABLED=0 go install github.com/alex123012/dependency-bot/cmd/dependency-bot
clean:
	rm -f $$GOPATH/bin/dependency-bot
	rm -f bin/*
