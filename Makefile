GOCMD=go
GOBUILD=$(GOCMD) build

bin/bindown:
	./script/bootstrap-bindown.sh -b bin

bin/gobin: bin/bindown
	bin/bindown download $@

bin/golangci-lint: bin/bindown
	bin/bindown download $@

.PHONY: clean
clean:
	rm -rf ./bin
