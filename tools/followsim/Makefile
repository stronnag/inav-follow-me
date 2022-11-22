APP = followsim
prefix ?= $$HOME/.local

$(APP):	$(wildcard *.go) go.sum
	go build -o $(APP) -ldflags "-w -s"

go.sum: go.mod $(wildcard *.go)
	go mod tidy

clean:
	@go clean
	@rm -f go.sum

install: $(APP)
	-install -d $(prefix)/bin
	-install -s $(APP) $(prefix)/bin/$(APP)
