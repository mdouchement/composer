# Composer

Composer is CLI software for managing processes in development environment.

## Requirements

- Golang 1.6.x/1.7.x

## Installation

- Download it from [releases](https://github.com/mdouchement/composer/releases) page.

### Manual build

1. Install Go 1.6 or 1.7
2. Install Glide dependency manager
  - `GO15VENDOREXPERIMENT=1 go get -u github.com/Masterminds/glide`
3. Clone this project
  - `git clone https://github.com/mdouchement/composer`
4. Fetch dependencies
  - `glide install`
5. Build the binary
  - `go build -o composer *.go`
6. Install the compiled binary
  - `mv composer /usr/bin/composer`

## Usage

```sh
$ composer start -c ~/server-stack.yml
```

## Configuration file

- Basic

```yml
services:
  risuto:
    pwd: $GOPATH/src/github.com/mdouchement/risuto
    command: go run risuto.go -p 5000 -b localhost
    environment:
      RISUTO_DATABASE: /tmp/data/tiedot_db

  breizhtrip:
    pwd: $GOPATH/src/github.com/mdouchement/breizhtrip-go
    command: go run breizhtrip.go -p 5005 -b localhost

  machinery:
    pwd: $GOPATH/src/github.com/mdouchement/machnery-app
    command: go run app.go worker -c 5
```

- Full options

```yml
settings:
  logger:
    buffer_size: 142 # Queue size of entries to be logged.
    entry_max_size: 865536 # Max length of an entry message. Default 65536 bytes.

services:
  ggpull:
    pwd: /home/$USER/myapp
    command: git pull

  app:
    hooks:
      wait:
        - ggpull
    pwd: /home/$USER/myapp
    command: bundle exec rails s
    log_trim_pattern: '\[.*\] \w+  (?P<message>.*)'
    environment:
      RAILS_ENV: production

  worker:
    hooks:
      wait: # wait for other processes before start
        - ggpull
      kill: # kill other processes on exit
        - app
    pwd: /home/$USER/myapp
    command: bundle exec sidekiq -c config/sidekiq.yml
```

### Trim logs

Outputed logs can be trimed to remove useless data like timestamp. This option is based on the Golang's [regexp](https://golang.org/pkg/regexp/) package and you can test your regexp with the following website [Regex Tester - Golang](https://regex-golang.appspot.com/assets/html/index.html).

The captured group must be named `message` like the following example:

```
log_trim_pattern: '\[.*\] \w+  (?P<message>.*)'
```

The result is the following:

```
# Without trim:
[Feb 13 10:38:09]  INFO my_app: [2017-02-13 10:38:09] INFO  WEBrick 1.3.1

# With trim:
[Feb 13 10:38:09]  INFO my_app: WEBrick 1.3.1

# When the trim regexp does not match, the message is prefixed with `[!]`
[Feb 13 10:38:09]  INFO my_app: [!] => Booting WEBrick
```

## License

**MIT**

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
