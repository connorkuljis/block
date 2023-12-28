build:
	go build -o block

run: 
	./block

clean:
	rm ./block

dev:
	reflex -g '*.go' -d fancy make
