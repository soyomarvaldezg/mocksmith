binary := "mocksmith"
build_dir := "build"

default: build

build:
    mkdir -p {{build_dir}}
    go build -o {{build_dir}}/{{binary}} .

clean:
    rm -rf {{build_dir}}

install:
    go install .

test:
    go test ./...

vet:
    go vet ./...

run:
    go run .
