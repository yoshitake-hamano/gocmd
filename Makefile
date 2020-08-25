
.PHONY: all
all:
	go generate ./...
	go build -o bin ./...

.PHONY: test
test: all
	cd test; ./main.sh
