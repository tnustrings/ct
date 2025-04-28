name = "ct"

# get the version from github tag
# delete the v from the version tag cause python build seems to strip it as well
version = $(shell git tag | tail -1)


all: ct.deb

ct.deb: ../bin/ct
	# create a directory for the package and in it a DEBIAN directory with meta info.
	mkdir -p "ct-${version}"
	mkdir -p "ct-${version}/DEBIAN"

	# make the bin

	mkdir -p "ct-${version}/usr/local/bin"

	# put the ct executable in

	cp ../bin/ct "ct-${version}/usr/local/bin"

	# manually create a file ct-<version>/deb/DEBIAN/control and put in the package info.
	# build the package
	dpkg-deb --build "ct-${version}"


release: 
	make
	gh release create "${version}" "ct-${version}.deb"

release-update: # if an asset was already uploaded, delete it before uploading again
	gh release delete-asset "${version}" "ct-${version}.deb" -y
	make release


