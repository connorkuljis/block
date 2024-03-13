build:
	go build -o . ./... 

run: 
	./block

clean:
	rm ./block

install:
	sudo cp ./block /usr/local/bin/

all: build run

dev:
	reflex make 

