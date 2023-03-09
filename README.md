# golling - update golang to the latest version
For those of you who want to stay up to date, golling will install or update the latest version of golang on your system. So, let's begin with 'golling update'.

<blockquote class="twitter-tweet" data-lang="en" data-dnt="true" data-theme="dark"><p lang="en" dir="ltr">🥳 Go 1.20.2 and 1.19.7 are released!<br><br>🔐 Security: Includes a security fix for crypto/elliptic (CVE-2023-24532).<br><br>🔈 Announcement: <a href="https://t.co/0R5yvEp6JL">https://t.co/0R5yvEp6JL</a><br><br>⬇️ Download: <a href="https://t.co/CxvhilnLor">https://t.co/CxvhilnLor</a><a href="https://twitter.com/hashtag/golang?src=hash&amp;ref_src=twsrc%5Etfw">#golang</a> <a href="https://t.co/xlnHevMUGv">pic.twitter.com/xlnHevMUGv</a></p>&mdash; Go (@golang) <a href="https://twitter.com/golang/status/1633151903352627207?ref_src=twsrc%5Etfw">March 7, 2023</a></blockquote>

## How to install
### Use "go install"
If you does not have the golang development environment installed on your system, please install golang from the [golang official website](https://go.dev/doc/install).
```
 go install github.com/nao1215/golling@latest
```

### Use homebrew (aarch64)
```
$ brew tap nao1215/tap
$ brew install nao1215/tap/golling
```

## How to use
golling start update if golang is not up to date. By default, golling checks /usr/local/go. If golang is not on the system, golling install the latest golang in /usr/local/go.
```
$ sudo golling update
download go1.20.1.linux-amd64.tar.gz at current directory
Downloading...99886/99886 kB (100%)
[compare sha256 checksum]
 expect: 000a5b1fca4f75895f78befeb2eecf10bfff3c428597f3f1e69133b63b911b02
 got   : 000a5b1fca4f75895f78befeb2eecf10bfff3c428597f3f1e69133b63b911b02

backup original /usr/local/go as /usr/local/go.backup
start extract go1.20.1.linux-amd64.tar.gz at /usr/local/go
delete backup (/usr/local/go.backup)
delete go1.20.1.linux-amd64.tar.gz

success to update golang (version 1.20.1)
```

## Support OS
- Linux
- Mac

## Contributing / Contact
First off, thanks for taking the time to contribute! heart Contributions are not only related to development. For example, GitHub Star motivates me to develop!
  
If you would like to send comments such as "find a bug" or "request for additional features" to the developer, please use one of the following contacts.
- [GitHub Issue](https://github.com/nao1215/golling/issues)

## LICENSE
The golling project is licensed under the terms of [MIT LICENSE](./LICENSE).

## Another project: gup update binaries installed by "go install".
[gup](https://github.com/nao1215/gup) command update binaries installed by "go install" to the latest version. gup updates all binaries in parallel, so very fast. It also provides subcommands for manipulating binaries under $GOPATH/bin ($GOBIN). It is a cross-platform software that runs on Windows, Mac and Linux.
