bindata:
	go-bindata ${BINDATA_FLAGS} -nometadata -o statics/bindata.go -pkg=statics -ignore=bindata.go statics/*
	gofmt -w -s statics/bindata.go

install: bindata
	go install -v ./...

build-arm:
	env GOOS=linux GOARCH=arm go build -v ./cmd/...
