
.PHONY: all
all:
	go generate ./...
	go build -o bin ./...

.PHONY: test
test: all
	go test ./...
	cd test/scripts; ./testMain.sh
	cd test/cpputest; $(MAKE) test
