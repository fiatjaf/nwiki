nwiki: $(shell find . -name "*.go")
	CC=$$(which musl-gcc) go build -ldflags="-s -w" -o ./nwiki

dist: $(shell find . -name "*.go")
	mkdir -p dist
	CC=$$(which musl-gcc) gox -ldflags="-s -w" -osarch="windows/amd64 darwin/amd64 linux/386 linux/amd64 linux/arm freebsd/amd64" -output="dist/nwiki_{{.OS}}_{{.Arch}}"
