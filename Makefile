BINARY=bin/ezwg
GUI_BINARY=bin/ezwg-gui
CMD_PATH=./cmd/ezwg
GUI_CMD_PATH=./cmd/ezwg-gui

.PHONY: all build gui test clean dist install

all: build gui

build:
	@echo "==> Building $(BINARY)..."
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(BINARY) $(CMD_PATH)
	@echo "==> Build finished: $(BINARY)"

gui:
	@echo "==> Building GUI Client $(GUI_BINARY)..."
	@mkdir -p bin
	go build -ldflags="-s -w" -o $(GUI_BINARY) $(GUI_CMD_PATH)
	@if [ "$$(uname -s)" = "Darwin" ]; then \
		./scripts/package-mac.sh; \
	fi
	@echo "==> GUI Build finished."

test:
	@echo "==> Running unit tests..."
	go test -v ./internal/...

dist:
	@echo "==> Building distribution release archives..."
	./scripts/build.sh

install: build gui
	@echo "==> Installing to /usr/local/bin..."
	sudo cp $(BINARY) /usr/local/bin/ezwg
	sudo ln -sf /usr/local/bin/ezwg /usr/local/bin/ezwireguard
	@if [ -f $(GUI_BINARY) ]; then \
		sudo cp $(GUI_BINARY) /usr/local/bin/ezwg-gui; \
	fi
	@if [ "$$(uname -s)" = "Darwin" ] && [ -d "dist/Easy WireGuard.app" ]; then \
		echo "==> Installing Easy WireGuard.app to /Applications..."; \
		sudo rm -rf "/Applications/Easy WireGuard.app"; \
		sudo cp -r "dist/Easy WireGuard.app" /Applications/; \
	elif [ "$$(uname -s)" = "Linux" ] && [ -d "/usr/share/applications" ]; then \
		echo "==> Installing desktop launcher to /usr/share/applications..."; \
		sudo cp assets/ezwg.desktop /usr/share/applications/; \
	fi
	@echo "==> Installed successfully."

clean:
	@echo "==> Cleaning..."
	rm -rf bin dist
