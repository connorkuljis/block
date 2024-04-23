build:
	go build -o . ./... 

run: 
	./block-cli serve

clean:
	rm ./block

install:
	sudo cp ./block-cli /usr/local/bin/

all: build run

dev:
	reflex -s make all

