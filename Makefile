.PHONY: wasm server frontend build-all clean

wasm:
	mkdir -p frontend/public
	GOOS=js GOARCH=wasm go build -o frontend/public/main.wasm cmd/wasm/main.go
	rm -f frontend/public/wasm_exec.js
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" frontend/public/
	chmod +w frontend/public/wasm_exec.js

server:
	cd frontend && npm run build
	mkdir -p cmd/server/dist
	cp -r frontend/dist/* cmd/server/dist/
	go build -o server_bin cmd/server/main.go

frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

clean:
	rm -rf server_bin frontend/dist frontend/public/main.wasm frontend/public/wasm_exec.js
