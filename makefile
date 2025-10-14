all: bin/ct 

bin/ct: bin/ctmini cmd/ct/main.ct ct.go org.ct tex.ct conf/*
	# make a mini-ct that doesn't depend on ct code itsself, use it to build the ct code for the main program
	./bin/ctmini cmd/ct/main.ct; mv main.go cmd/ct # ctmini doesn't move outputfiles into their corresponding directories
	./bin/ctmini org.ct
	./bin/ctmini tex.ct
	go build -o bin/ct cmd/ct/main.go

bin/ctmini: # build ctmini
	go build -o bin/ctmini cmd/ctmini/main.go

.PHONY deb:
	cd deb; make
