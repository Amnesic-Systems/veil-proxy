prog = nitriding-proxy
deps = cmd/*.go *.go go.mod go.sum Makefile

$(prog): $(deps)
	go build -C cmd/ -o ../$(prog)

.PHONY: clean
clean:
	rm -f $(prog)
