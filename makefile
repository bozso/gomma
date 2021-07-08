.PHONY: build deploy

test:
	go test ./command ./meta ./stream

/tmp/gomma:
	go build -o /tmp/gomma

build: /tmp/gomma

lint:
	golangci-lint run

deploy: /tmp/gomma
	cp /tmp/gomma /home/istvan/mount/zafir/packages/usr/bin/gomma