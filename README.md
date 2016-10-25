# Composer

Composer is CLI software for managing processes in development environment.

## Requirements

- Golang 1.6.x

## Installation

- Download it from [releases](https://github.com/mdouchement/composer/releases) page.

### Manual build

1. Install Go 1.6
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
ggpull:
  pwd: /home/$USER/myapp
  command: git pull

app:
  hooks:
    wait:
      - ggpull
  pwd: /home/$USER/myapp
  command: bundle exec rails s
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

## License

**MIT**

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
