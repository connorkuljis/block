build:
	go build -o block

run: 
	./block

clean:
	rm ./block

release:
	sudo cp ./block /usr/local/bin/

dev:
	reflex -g '*.go' -d fancy make
