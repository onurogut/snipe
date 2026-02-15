BINARY=snipe
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-s -w"

.PHONY: build install clean

build:
	go build $(LDFLAGS) -o $(BINARY) .

install: build
	cp $(BINARY) /usr/local/bin/$(BINARY)

uninstall:
	rm -f /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
