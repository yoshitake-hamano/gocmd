
.PHONY: all
all:
	go build -o bin ./...

.PHONY: test
test: all
	cd test; ./main.sh
