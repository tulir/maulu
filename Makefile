build:
	go build -o maulu

package-prep: build
	cp maulu package/usr/bin/
	cp config.json redirect.html package/etc/maulu/

package: package-prep
	dpkg-deb --build package maulu.deb > /dev/null

clean:
	rm -f maulu maulu.deb package/usr/bin/maulu package/etc/maulu/redirect.html package/etc/maulu/config.json
