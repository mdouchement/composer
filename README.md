# Composer

Composer is CLI software for managing processes in development environment.

## Installation

- Download it from [releases](https://github.com/mdouchement/composer/releases) page.

### Manual build

`go install github.com/mdouchement/composer@latest`

or

1. Install Go 1.24+
2. Clone this project
  - `git clone https://github.com/mdouchement/composer`
3. Build the binary
  - `go build -o composer .`
4. Install the compiled binary
  - `mv composer /usr/bin/composer`

## Usage

```sh
$ composer start -c ~/server-stack.yml
```

## Configuration file

`command` is interpreted as shell script.

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

  random_shell_script:
    pwd: /tmp
    command: |
      for i in 0 1 2 3 4 5 6 7 8 9
      do
        echo "Hello $i times"
      done
```

- Full options

```yml
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
    ignore_error: true # Do not stop other services on error
    hooks:
      wait: # wait for other processes before start
        - ggpull
      kill: # kill other processes on exit
        - app
    pwd: /home/$USER/myapp
    command: bundle exec sidekiq -c config/sidekiq.yml
```

### Trim logs

Outputed logs can be trimed to remove useless data like timestamp. This option is based on the Golang's [regexp](https://golang.org/pkg/regexp/) package and you can test your regexp with the following website [regex101 with Golang flavor](https://regex101.com).

All the captured groups are printed by composer like the following example:

```
log_trim_pattern: '\[[^]]+\] \w+ (.*)'
```

The result is the following:

```
# Without trim:
my_app: [2017-02-13 10:38:09] INFO  WEBrick 1.3.1

# With trim:
my_app: WEBrick 1.3.1
```
> Nothing happens if the regex is invalid.\
> If the logs are colored you have to add ANSI colors like `\x1B[32m INFO \x1B[0m` in your regexp escaped as `log_trim_pattern: "\x1B\\[32m INFO \x1B\\[0m"`.

You can add a composer log file in order to see the ANSI colors escape sequences:

```yaml
settings:
  log_file: composer.log
```

## License

**MIT**

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
