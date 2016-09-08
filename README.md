![HubGraph](https://raw.githubusercontent.com/nmaggioni/HubGraph/master/banner.png)

[![standard-readme compliant](https://img.shields.io/badge/standard--readme-OK-green.svg?style=flat-square)](https://github.com/RichardLitt/standard-readme)

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

Given that your `$PATH` already has `$GOPATH/bin` in it, get the package and install it these commands:

```
$ go get github.com/nmaggioni/hubgraph
$ cd $GOPATH/src/github.com/nmaggioni/hubgraph
$ ./build.sh
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

PRs gladly accepted!

Small note: If editing the Readme, consider conforming to the [standard-readme](https://github.com/RichardLitt/standard-readme) specification.

## License

[MIT © Niccolò Maggioni](https://github.com/nmaggioni/HubGraph/blob/master/LICENSE)
