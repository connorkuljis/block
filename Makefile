build:
	go build -o block

run: 
	./block

clean:
	rm ./block

install:
	sudo cp ./block /usr/local/bin/

all: build run

dev:
	reflex make 

