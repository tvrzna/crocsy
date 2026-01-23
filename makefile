DIST_FILE=crocsy
BUILD_VERSION=`git describe --tags`

clean:
	rm -r dist

build:
	mkdir -p dist
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.buildVersion=${BUILD_VERSION}" -trimpath -o dist/${DISTFILE} -buildvcs=false

install:
	install -DZs dist/${DISTFILE} ${DESTDIR}/usr/bin