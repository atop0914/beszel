.PHONY: build run test lint clean

BINARY := beszel
CMD := ./cmd/beszel

build:
	go build -o $(BINARY) $(CMD)

run: build
	./$(BINARY) serve

test:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)
	rm -f beszel.db
