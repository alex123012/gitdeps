.PHONY: build install clean

versionflags?=
module?=github.com/alex123012/gitdeps/cmd/main
binary?=bin/gitdeps
extldflags?=
ldflags?=
_flags=-v -a -tags netgo -ldflags="-extldflags '-static $(extldflags)' -s -w $(versionflags) $(ldflags)"

build:
	CGO_ENABLED=0 go build $(_flags) -o $(binary) $(module)
# build-race:
# 	CGO_ENABLED=1 go build $(_flags) -race -o $(binary) $(module)
# install:
# 	CGO_ENABLED=0 go install $(module)
clean:
	rm -f $$GOPATH/$(binary)
	rm -f bin/*
