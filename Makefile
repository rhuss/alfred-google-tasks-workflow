BINARY := gtasks-bin
LAUNCHER := gtasks
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
	cp $(LAUNCHER) $(BUILD_DIR)/workflow/
	cp info.plist $(BUILD_DIR)/workflow/
	cp icon.png $(BUILD_DIR)/workflow/
	cp icons/*.png $(BUILD_DIR)/workflow/icons/
	cd $(BUILD_DIR)/workflow && zip -r ../../$(WORKFLOW) .
	@echo "Created $(WORKFLOW)"

install: package
	open $(WORKFLOW)

release:
	@if [ -z "$(VERSION)" ]; then echo "Usage: make release VERSION=v1.2.0"; exit 1; fi
	@# Update version in info.plist (strip leading 'v' from tag)
	sed -i '' 's|<key>version</key>|<key>version</key>|; \
		/<key>version<\/key>/{n;s|<string>[^<]*</string>|<string>'"$${VERSION#v}"'</string>|;}' info.plist
	git add info.plist
	git commit -m "release: $(VERSION)" --allow-empty
	$(MAKE) package
	git tag $(VERSION)
	git push origin main $(VERSION)
	gh release create $(VERSION) $(WORKFLOW) \
		--title "$(VERSION)" \
		--generate-notes
