bin = ./bin/

build:
	go build -o $(bin) .

run: 
	$(bin)block serve

clean:
	rm $(bin)block

install:
	sudo cp ./bin/block/ /usr/local/bin/block

all: build run

dev:
	reflex -s make all

