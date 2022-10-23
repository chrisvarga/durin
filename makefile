durin:
	go build durin.go

all: durin

install:
	cp durin /usr/local/bin

clean:
	rm durin
