# Define the Go compiler and flags
GO := go
GOFLAGS :=

# Define the target executable name
TARGET := tandum

# Define the source files
SOURCES := $(wildcard *.go)

# Define the build rule
build:
	$(GO) build $(GOFLAGS) -o $(TARGET) $(SOURCES)

# Define the clean rule
clean:
	rm -f $(TARGET)