bin = ./bin/
exec = block

build:
	go build -o $(bin)$(exec) .

run: 
	$(bin)$(exec)

clean:
	rm $(bin)$(exec)

install:
	sudo cp $(bin)$(exec) /usr/local/bin/$(exec)

all: build run

dev:
	reflex -s make all

