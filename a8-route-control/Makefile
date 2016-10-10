.PHONY: run clean

SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
EXE=a8routing
IMAGE=weavescope-a8routing2-plugin
UPTODATE=.$(EXE).uptodate

run: $(UPTODATE)
	# Create a socket for communicating IP addresses of other containers on the host to this container
	mkdir -p /tmp/dockerConnection;
	touch /tmp/dockerConnection/hostconnection.sock
	python hostDockerConnector.py &
	# --net=host gives us the remote hostname, in case we're being launched against a non-local docker host.
	# We could also pass in the `-hostname=foo` flag, but that doesn't work against a remote docker host.
	docker build -t $(IMAGE) .
	$(SUDO) docker run --rm -it \
		--net=host \
		-v /var/run/scope/plugins:/var/run/scope/plugins \
		-v /tmp/dockerConnection:/var/run/dockerConnection \
		--name $(IMAGE) $(IMAGE)

$(UPTODATE): $(EXE) Dockerfile
	$(SUDO) docker build -t $(IMAGE) .
	touch $@

$(EXE): main.go
	$(SUDO) docker run --rm -v "$$PWD":/usr/src/$(EXE) -w /usr/src/$(EXE) golang:1.6 go build -v

clean:
	- rm -rf $(UPTODATE) $(EXE)
	- $(SUDO) docker rmi $(IMAGE)
