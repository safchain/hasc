.DEFAULT_GOAL := bindata

.PHONY: app
app: 
	cd app && npm run build && cd ..
	rm -rf statics/*
	cp -R app/build/* statics/

bindata: app
	go-bindata ${BINDATA_FLAGS} -nometadata -o statics/bindata.go \
		-pkg=statics -ignore=bindata.go statics/* statics/static/* statics/static/css/* statics/static/js/*
	gofmt -w -s statics/bindata.go

install: bindata
	go install -v ./...

build-arm:
	env GOOS=linux GOARCH=arm go build -v ./cmd/...
