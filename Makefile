build:
	go build -o block

run: 
	./block

clean:
	rm ./block

install:
	sudo cp ./block /usr/local/bin/

dev:
	go build -o block
	reflex -r '\.go$\' make

