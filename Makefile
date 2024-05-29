# Variables
IMAGE_NAME := dimutils
IMAGE_TAG := 0.1
REGISTRY := registry.hub.docker.com/karanko

# Build the container
build:
	podman build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# Push the container to the registry
push:
	podman push $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Clean up the container
clean:
	podman rmi $(IMAGE_NAME):$(IMAGE_TAG)

# Default target
all: build
