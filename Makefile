BINARY := gtasks
BUNDLE_ID := com.rhuss.gtasks
WORKFLOW := alfred-google-tasks.alfredworkflow
BUILD_DIR := build

.PHONY: build test clean package install release

build:
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY) ./cmd/
	codesign -s - $(BUILD_DIR)/$(BINARY)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(WORKFLOW)

package: build
	@mkdir -p $(BUILD_DIR)/workflow/icons
	cp $(BUILD_DIR)/$(BINARY) $(BUILD_DIR)/workflow/
	cp info.plist $(BUILD_DIR)/workflow/
	cp icon.png $(BUILD_DIR)/workflow/
	cp icons/*.png $(BUILD_DIR)/workflow/icons/
	cd $(BUILD_DIR)/workflow && zip -r ../../$(WORKFLOW) .
	@echo "Created $(WORKFLOW)"

install: package
	xattr -cr $(WORKFLOW)
	open $(WORKFLOW)

release: package
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=v1.1.0"; exit 1; fi
	git tag $(VERSION)
	git push origin $(VERSION)
	gh release create $(VERSION) $(WORKFLOW) \
		--title "$(VERSION)" \
		--generate-notes
