build:
	GOOS=linux GOARCH=amd64 go build -v -o bin/listbot .

install:
	glide install

heroku:
	build

all: install build
