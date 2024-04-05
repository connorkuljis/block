build:
	go build -o . ./... 

run: 
	./block serve

clean:
	rm ./block

install:
	sudo cp ./block /usr/local/bin/

all: build run

dev:
	reflex -s make all

