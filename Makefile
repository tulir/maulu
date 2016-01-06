maulu: $(shell ls *.go)
	go install

compile: $(shell ls *.go)
	env GOOS=linux GOARCH=amd64 go build -v

package: compile maulu $(shell ls *.html)
	tar cvfJ maulu.tar.xz maulu *.html

maumain: package
	scp maulu.tar.xz main:Maulu/

production: maumain clean

clean:
	rm -f maulu maulu.tar.xz
