#! /bin/sh

$GOPATH/bin/rice append --exec $(go build -v 2>&1 | cut -d/ -f3)