# name
name = "ct"

# get the version from github tag
# sort by version; get the last line; delete the v from the version tag cause python build seems to strip it as well
version = $(shell git tag | sort -V | tail -1 | tr -d v)


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

publish-update: # if an asset was already uploaded, delete it before uploading again
	make
	# does the tag updating also update the source code at the resource?
	# move the version tag to the most recent commit
	git tag -f "v${version}"
	# delete tag on remote
	git push origin ":refs/tags/v${version}" 
	# re-push the tag to the remote
	git push --tags
