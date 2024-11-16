prog = veil-proxy
deps = cmd/*.go *.go go.mod go.sum Makefile

all: cap

$(prog): $(deps)
	go build -C cmd/ -o ../$(prog)

.PHONY: cap
cap: $(prog)
	sudo setcap cap_net_admin=ep $(prog)

.PHONY: clean
clean:
	rm -f $(prog)
