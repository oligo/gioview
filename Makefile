# https://github.com/oligo/gioview

NAME=gioview

BIN_ROOT=$(PWD)/.bin
export PATH:=$(PATH):$(BIN_ROOT)

print:

bin:
	mkdir -p $(BIN_ROOT)
	cd example/basic && go build -o $(BIN_ROOT)/$(NAME) .
run:
	$(NAME)