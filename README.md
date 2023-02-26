# golling - update golang to the latest version
For those of you who want to stay up to date, golling will install or update the latest version of golang on your system. So, let's begin with 'golling update'.


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
sudo golling update
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