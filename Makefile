VERSION=1.0

build-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -o build/guardian .
	tar cfvz "build/guardian-${VERSION}-linux-amd64.tar.gz" -C build guardian
	rm build/guardian
