default: out/example

clean:
	rm -rf out

test:
	go test -timeout 30s github.com/fujivara/lab3-arcprog/painter

out/example:
	mkdir -p out
	go build -o out/example ./cmd/painter/main.go