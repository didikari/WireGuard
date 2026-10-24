BINARY=bin/ezwg
CMD_PATH=./cmd/ezwg

.PHONY: all build test clean dist install

all: build

build:
	@echo "==> Building $(BINARY)..."
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) $(CMD_PATH)
	@echo "==> Build finished: $(BINARY)"

test:
	@echo "==> Running unit tests..."
	go test -v ./...

dist:
	@echo "==> Building distribution release archives..."
	./scripts/build.sh

install: build
	@echo "==> Installing to /usr/local/bin/ezwg..."
	sudo cp $(BINARY) /usr/local/bin/ezwg
	sudo ln -sf /usr/local/bin/ezwg /usr/local/bin/ezwireguard
	@echo "==> Installed successfully."

clean:
	@echo "==> Cleaning..."
	rm -rf bin dist
