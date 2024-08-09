# Variables
IMAGE_NAME := dimutils
IMAGE_TAG := 0.2
REGISTRY := docker.io/dim9

# Build the container
build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
# tag the container to the registry
tag: build
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):latest

# Push the container to the registry
push: tag
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):latest

# Clean up the container
clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG)

run:
	docker run -it --rm $(IMAGE_NAME):$(IMAGE_TAG)

# Default target
all: build
