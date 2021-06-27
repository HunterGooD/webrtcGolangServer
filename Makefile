
build: clean
	cd ./cmd/conference && \
	go build -o ../../dist/server-sfu

clean:
	@rm -rf ./dist

certs:
	mkdir configs/certs
	openssl genrsa -out configs/certs/server.key 2048
	openssl ecparam -genkey -name secp384r1 -out configs/certs/server.key
	openssl req -new -x509 -sha256 -key configs/certs/server.key -out configs/certs/server.crt -days 3650

clear_certs:
	rm -rf configs/certs