MAIN := gamma

$(MAIN): gamma.go
	go build gamma.go

deploy: $(MAIN)
	scp $(MAIN) istvan@zafir.ggki.hu:/home/istvan/packages/usr/bin
