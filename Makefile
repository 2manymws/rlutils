export GO111MODULE=on

default: test

ci: depsdev test

test:
	go test ./... -coverprofile=coverage.out -covermode=count

lint:
	golangci-lint run ./...
	go vet -vettool=`which gostyle` -gostyle.config=$(PWD)/.gostyle.yml ./...

depsdev:
	go install github.com/Songmu/ghch/cmd/ghch@latest
	go install github.com/Songmu/gocredits/cmd/gocredits@latest
	go install github.com/k1LoW/octocov-go-test-bench/cmd/octocov-go-test-bench@latest
	go install github.com/k1LoW/octocov-cachegrind/cmd/octocov-cachegrind@latest
	go install github.com/k1LoW/gostyle@latest

prerelease:
	git pull origin main --tag
	go mod download
	ghch -w -N ${VER}
	gocredits -w .
	cat _EXTRA_CREDITS >> CREDITS
	git add CHANGELOG.md CREDITS go.mod go.sum
	git commit -m'Bump up version number'
	git tag ${VER}

prerelease_for_tagpr: depsdev
	go mod download
	gocredits -w .
	cat _EXTRA_CREDITS >> CREDITS
	git add CHANGELOG.md CREDITS go.mod go.sum

release:
	git push origin main --tag
