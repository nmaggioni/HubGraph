#! /bin/sh

echo -n "Do you want to cross-compile? [Y/N] "
read YN
if [ "$YN" == "Y" ] || [ "$YN" == "y" ]; then
    echo -e "\n* Preparing asset bundle"
    $GOPATH/bin/rice embed-go
    echo "* Compiling for Linux x86_64"
    GOOS=linux GOARCH=amd64 go build -o dist/hubgraph_linux_amd64.bin
    echo "* Compiling for Linux x86"
    GOOS=linux GOARCH=386 go build -o dist/hubgraph_linux_i386.bin
    echo "* Compiling for ARMv6"
    GOARCH=arm GOARM=6 go build -o dist/hubgraph_armv6.bin
    echo "* Compiling for ARMv5"
    GOARCH=arm GOARM=5 go build -o dist/hubgraph_armv5.bin
    echo "* Compiling for Windows x86_64"
    GOOS=windows GOARCH=amd64 go build -o dist/hubgraph_windows_amd64.exe
    echo "* Compiling for Windows x86"
    GOOS=windows GOARCH=386 go build -o dist/hubgraph_windows_i386.exe
    echo "* Compiling for Mac OS x86_64"
    GOOS=darwin GOARCH=amd64 go build -o dist/hubgraph_darwin_amd64
    rm rice-box.go
    file dist/*
else
    echo -e "\n* Compiling and appending bundle"
    $GOPATH/bin/rice append --exec $(go build -v 2>&1 | cut -d/ -f3)
    file hubgraph
fi

