.PHONY: all clean

# Compiler and flags
GO := go
GOFLAGS := -v

# Output binary name
OUTPUT := unexpect

# Source files
SOURCES := main.go

# Build target
all: $(OUTPUT)

$(OUTPUT): $(SOURCES)
	$(GO) build $(GOFLAGS) -o $(OUTPUT) $(SOURCES)
run: $(OUTPUT)
	./$(OUTPUT)

tidy:
	$(GO) mod tidy
# Clean target
clean:
	rm -f $(OUTPUT)

tests: func-tests

func-tests:
	@cd tests && ../$(OUTPUT)