all: ct.deb

ct.deb: ../bin/ct
	# create a directory for the package and in it a DEBIAN directory with meta info.
	mkdir -p ct
	mkdir -p ct/DEBIAN

	# make the bin

	mkdir -p ct/usr/local/bin

	# put the ct executable in

	cp ../bin/ct ct/usr/local/bin

	# manually create a file deb/DEBIAN/control and put in the package info.
	# build the package
	dpkg-deb --build ct
