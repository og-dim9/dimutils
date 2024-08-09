# Variables
IMAGE_NAME := dimutils
IMAGE_TAG := 0.1.2
REGISTRY := docker.io/dim9

# Build the container
build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	docker build -f ./Dockerfile.thick -t $(IMAGE_NAME):$(IMAGE_TAG)-thick .
# tag the container to the registry
tag: build
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-thin
	docker tag $(IMAGE_NAME):$(IMAGE_TAG)-thick $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-thick
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):latest
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):latest-thin
	docker tag $(IMAGE_NAME):$(IMAGE_TAG)-thick $(REGISTRY)/$(IMAGE_NAME):latest-thick
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):thin
	docker tag $(IMAGE_NAME):$(IMAGE_TAG)-thick $(REGISTRY)/$(IMAGE_NAME):thick

# Push the container to the registry
push: tag
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-thin
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)-thick
	docker push $(REGISTRY)/$(IMAGE_NAME):latest
	docker push $(REGISTRY)/$(IMAGE_NAME):latest-thin
	docker push $(REGISTRY)/$(IMAGE_NAME):latest-thick
	docker push $(REGISTRY)/$(IMAGE_NAME):thin
	docker push $(REGISTRY)/$(IMAGE_NAME):thick

# Clean up the container
clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG)
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG)-thick

run:
	docker run -it --rm $(IMAGE_NAME):$(IMAGE_TAG)

# Default target
all: build
