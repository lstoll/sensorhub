export GO15VENDOREXPERIMENT = 1
OUT := out
GOCMD := go
GOBUILD := $(GOCMD) build
GOINSTALL := $(GOCMD) install
PACKAGES := $(shell go list ./... | grep -v /vendor/)

CMD_LIST := itron-sdr

BUILD_LIST = $(foreach int, $(CMD_LIST), $(int)_build)
CLEAN_LIST = $(foreach int, $(CMD_LIST), $(int)_clean)

.PHONY: govendor $(CMD_LIST) $(BUILD_LIST) $(CLEAN_LIST) build package out

all: build package

out:
	mkdir -p $(OUT)

build: out $(BUILD_LIST)
clean: $(CLEAN_LIST)

$(CMD_LIST):
	cd $@ && $(GOBUILD)

$(BUILD_LIST): %_build:
	$(GOBUILD) -o out/$* ./$*
$(CLEAN_LIST): %_clean:
	rm -rf $(OUT)
	rm -fr stage
	rm *.deb

package: build
	mkdir -p stage/usr/bin/
	cp $(OUT)/* stage/usr/bin/
	fpm -s dir -t deb -C stage \
		-n sensorhub \
		-v $(shell git describe --abbrev=0 --tags) \
		-d "rtlamr > 0" \
		usr
	rm -fr stage

govendor:
	$(GOINSTALL) github.com/kardianos/govendor
