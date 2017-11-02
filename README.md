# which-better

Help you to choose the better GitHub repositories.

Rate score by github stars, last push date and the number of contributors.

Note: more smaller score, more better.

## Installation

Via go-get or the binary [releases](https://github.com/songjiayang/which-better/releases).

```
go get github.com/songjiayang/which-better
```

## Usage

* Add an alias: `alias whb=which-better`
* Help text:
```
  $ which-better -h
  Usage:
    which-better -r xxx,xxx
    which-better -v

  Options:
    -r the compared repositories split by ","
    -v dispaly version
```
* Examples:
```
$ which-better -r golang/dep,Masterminds/glide,tools/godep,kardianos/govendor,pote/gpm

  golang/dep: 4
  Masterminds/glide: 5
  tools/godep: 11
  kardianos/govendor: 11
  pote/gpm: 14
```
For this example you can see `golang/dep` and `Masterminds/glide` are better.

## Changelog

### 0.0.1

Initial release.

## License

MIT
