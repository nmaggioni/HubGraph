![HubGraph](https://raw.githubusercontent.com/nmaggioni/HubGraph/master/banner.png)

[![Codacy Badge](https://api.codacy.com/project/badge/Grade/b767236ba1f348d08b93ed317d657ed3)](https://www.codacy.com/app/nmaggioni/HubGraph?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=nmaggioni/HubGraph&amp;utm_campaign=Badge_Grade) [![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

> Grab the latest events from the boring GitHub's API and build an entertaining graph upon them!


## Table of Contents

- [Install](#install)
	- [Packaged releases](#packaged-releases)
	- [From source](#from-source)
- [Usage](#usage)
- [Examples](#examples)
- [Contribute](#contribute)
- [License](#license)

## Install

### Packaged releases

Check out [the releases section](https://github.com/nmaggioni/HubGraph/releases) for ready-to-run binaries, with all the needed dependencies already embedded. [Here's the latest one!](https://github.com/nmaggioni/HubGraph/releases/latest)

### From source

Please ensure that you have [go.rice](https://github.com/GeertJohan/go.rice) installed. See [here](https://github.com/GeertJohan/go.rice#installation) for the official commands.

Given that your `$PATH` already has `$GOPATH/bin` in it, get the package and install it these commands:

```
$ go get github.com/nmaggioni/hubgraph
$ cd $GOPATH/src/github.com/nmaggioni/hubgraph
```
Now use the `build.sh` script if you want to cross-compile, or just run the following to build a binary for your system:

```
rice append --exec $(go build -v 2>&1 | cut -d/ -f3)
```
The last step is to install the binary to the `$GOPATH/bin` directory:

```
$ go install
```

## Usage

HubGraph has some useful command line options, you can check them by using the help flag:

```
$ ./hubgraph -h
```

## Examples

Here are three examples of what HubGraph will produce: the _blue_ points are repositories, and other coloured nodes linked to them are related events. _Dark green_, for example, is for when an issue has been commented, _yellow_ for a new commit pushed, _light blue_ for new PRs submitted, and so on... **Place the mouse over a node to read its description!**

![HubGraph](https://raw.githubusercontent.com/nmaggioni/HubGraph/master/demo.png)

## Contribute

PRs gladly accepted! Basing them on a new feature/fix branch would help in reviewing.

Small note: If editing the Readme, consider conforming to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

[MIT © Niccolò Maggioni](https://github.com/nmaggioni/HubGraph/blob/master/LICENSE)
