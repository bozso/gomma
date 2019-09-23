MAIN := gamma

SRC := $(wildcard ../gamma/*.go)

$(MAIN): gamma.go $(SRC)
	go build gamma.go

deploy: $(MAIN)
	scp $(MAIN) istvan@zafir.ggki.hu:/home/istvan/packages/usr/bin
