.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	go-assets-builder -o config/billing/bindata.go -p billing billing/servicename.yml
	env GOOS=linux go build -ldflags="-s -w" -o release/bin/billing billing/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh
