APP_NAME := gitaskop

ifeq ($(OS),Windows_NT)     # is Windows_NT on XP, 2000, 7, Vista, 10...
    detected_OS := windows
else
    detected_OS := $(shell uname  | tr A-Z a-z)  # same as "uname -s"
endif

# Go compiler
GO := go

# Build targets
.PHONY: all linux windows clean docker run test

default: $(detected_OS)

all: linux windows

dl:
	@$(GO) mod download

linux: dl
	@GOOS=linux $(GO) build -o $(APP_NAME) *.go

windows: dl
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(APP_NAME).exe *.go

docker:
	@docker build -t dim9/dimutils:gitaskop .
docker-push: docker
	@docker push dim9/dimutils:gitaskop
docker-run: docker
	@docker run -it --rm dim9/dimutils:gitaskop

run: run-$(detected_OS)

run-linux: linux
	@./$(APP_NAME)

run-windows: windows
	@$(APP_NAME).exe

clean:
	@echo "Cleaning..."
	@rm -fv $(APP_NAME) $(APP_NAME).exe
	@make -C test clean

test:
	@cd test && make clean curl


watch:
	@fswatch -o . | xargs -n1 -I{} make run