.PHONY : tools mock docs

all: example

mod:
	go mod tidy

example: mod
	go build -o example/example ./example/*.go
