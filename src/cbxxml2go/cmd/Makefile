# LDFLAGS = -ldflags "-linkmode external -extldflags -static"

build:
	go build -o cbxml2x ${LDFLAGS} *.go

run: build
	./cbxml2x --xml ../test/examples/O700010F.xml