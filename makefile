all: bin/ct 

bin/ctmini: ctmini/main.go ct/ct.go
	# build without org and tex code that comes from ct files
	rm -f ct/tex.go ct/org.go # -f: don't give an error message
	cd ct; go build -o ../bin/ctmini

bin/ct: main.ct ct/ct.go ct/org.ct ct/tex.ct fc/fc.ct
	./bin/ctmini main.ct
	cd ct; ../bin/ctmini org.ct; ../bin/ctmini tex.ct
	cd ct; go build; cd ..
	# cd ctmini; go build; cd ..
	go build -o bin/ct

.PHONY deb:
	cd deb; make
