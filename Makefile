prog = nitriding-proxy
deps = *.go go.mod go.sum Makefile

$(prog): $(deps)
	go build -o $(prog)

.PHONY: clean
clean:
	rm -f $(prog)
