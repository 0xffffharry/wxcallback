NAME = wxcallback
PARAMS = -v -trimpath -ldflags "-s -w -buildid="
MAIN = ./cmd/wxcallback

build:
	go build -o wxcallback $(PARAMS) $(MAIN)

install:
	go install $(PARAMS) $(MAIN)
