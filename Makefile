linux:
   GOOS=linux GOARCH=amd64 go build -v -ldflags "-X main.version=`cat VERSION` -X main.branch=`git rev-parse --abbrev-ref HEAD` -X main.revision=`git rev-parse --short HEAD`"
mac:
   GOOS=darwin GOARCH=amd64 go build
