build:
	go build -o . ./... 

run: 
	./block-cli serve

clean:
	rm ./block

install:
	sudo cp ./block /usr/local/bin/block

all: build run

dev:
	reflex -s make all

