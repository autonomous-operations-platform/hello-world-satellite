BIN_DIR     := bin
PLATFORMS   := linux/amd64 linux/arm64

# OCM component constructor and derived component name.
CONSTRUCTOR := .ocm/component-constructor.yaml
COMPONENT   := $(shell sed -n 's/^  - name: \(.*\)$$/\1/p' $(CONSTRUCTOR) 2>/dev/null | head -n 1)

# Signing is optional. A fresh public clone has no key, so `make bundle`
# produces an unsigned CTF. Provide a key to sign: `make bundle-signed`.
SIGN_KEY    ?= /tmp/satellite-sign.key
PUB_KEY     := .ocm/satellite.pub.pem
SIG_NAME    := satellite-sig

REGISTRY          ?= autonomous-operations-docker-releases.common.repositories.cloud.sap
GENERIC_REGISTRY  ?= autonomous-operations-generic-releases
ARTIFACTORY_BASE  := https://common.repositories.cloud.sap/artifactory

# Derive VERSION from the exact git tag on HEAD (X.Y.Z).
# Override for local testing: make bundle VERSION=0.1.0
# Note: an override passes preflight but not check-publishable.
VERSION     ?= $(shell git describe --tags --exact-match HEAD 2>/dev/null | sed 's/^v//')

# Name of this satellite, derived from the repo directory. Used for binary and
# archive names. The constructor resource paths must match: binaries are built
# to $(BIN_DIR)/$(NAME)-<os>-<arch>.
NAME        := hello-world-satellite

CTF_DIR     = $(BIN_DIR)/ctf
CV_REF      = ctf::./$(CTF_DIR)//$(COMPONENT):$(VERSION)
AIRGAP_TGZ  = $(BIN_DIR)/$(NAME)-$(VERSION).ctf.tgz

.PHONY: build clean test \
    preflight compile package bundle sign verify bundle-signed \
    check-publishable bundle-airgap publish-registry publish-airgap publish

# ---------------------------------------------------------------------------
# Development targets
# ---------------------------------------------------------------------------

# Build the satellite binary for the current platform.
build:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -trimpath -o $(BIN_DIR)/$(NAME) .

clean:
	rm -rf $(BIN_DIR)/

test:
	go test ./...

# ---------------------------------------------------------------------------
# OCM bundle
# ---------------------------------------------------------------------------
# `make bundle` compiles all platforms and packages an unsigned CTF.
# `make bundle-signed` additionally signs and verifies (requires a signing key
# and $(PUB_KEY); see the signing config in .ocm/config).

preflight:
	@echo "==> preflight checks"
	@command -v ocm >/dev/null 2>&1 || { echo "ERROR: ocm CLI not found in PATH"; exit 1; }
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: HEAD is not tagged with a vX.Y.Z tag (shallow clone? run: git fetch --tags)"; \
		echo "       or override for local testing: make bundle VERSION=0.1.0"; \
		exit 1; \
	fi
	@if ! echo "$(VERSION)" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$$'; then \
		echo "ERROR: VERSION '$(VERSION)' is not valid semver (major.minor.patch)"; \
		exit 1; \
	fi
	@test -f "$(CONSTRUCTOR)" || { echo "ERROR: constructor not found at $(CONSTRUCTOR)"; exit 1; }
	@test -n "$(COMPONENT)"   || { echo "ERROR: could not read component name from $(CONSTRUCTOR)"; exit 1; }
	@for entry in $(PLATFORMS); do \
		os=$${entry%/*}; arch=$${entry#*/}; \
		if ! grep -qE "path: \./$(BIN_DIR)/$(NAME)-$$os-$$arch *$$" $(CONSTRUCTOR); then \
			echo "ERROR: platform $$entry is in PLATFORMS but $(CONSTRUCTOR) has no resource with path ./$(BIN_DIR)/$(NAME)-$$os-$$arch"; \
			exit 1; \
		fi; \
	done
	@have=$$(grep -c "path: \./$(BIN_DIR)/$(NAME)-" $(CONSTRUCTOR)); \
	if [ "$$have" -ne $(words $(PLATFORMS)) ]; then \
		echo "ERROR: $(CONSTRUCTOR) declares $$have platform resources but PLATFORMS lists $(words $(PLATFORMS))"; \
		exit 1; \
	fi
	@echo "    component: $(COMPONENT)"
	@echo "    version:   $(VERSION)"
	@echo "    platforms: $(PLATFORMS)"

compile: preflight
	@echo "==> compiling $(NAME) for $(PLATFORMS)"
	@mkdir -p $(BIN_DIR)
	@set -e; for entry in $(PLATFORMS); do \
		os=$${entry%/*}; arch=$${entry#*/}; \
		out=$(CURDIR)/$(BIN_DIR)/$(NAME)-$$os-$$arch; \
		echo "    $$os/$$arch -> $$out"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch go build -trimpath -o $$out .; \
	done

package: compile
	@echo "==> packaging CTF"
	@rm -rf $(CTF_DIR)
	@mkdir -p $(CTF_DIR)
	VERSION=$(VERSION) ocm add component-version \
		--working-directory . \
		--constructor $(CONSTRUCTOR) \
		--repository ctf::./$(CTF_DIR)
	@echo "==> descriptor"
	@ocm get component-version $(CV_REF) -o yaml

bundle: package
	@echo "==> unsigned bundle at $(CTF_DIR) ($(COMPONENT):$(VERSION), platforms: $(PLATFORMS))"
	@echo "    to sign: make bundle-signed"

# --- Optional signing -------------------------------------------------------

sign: package
	@test -f "$(PUB_KEY)"  || { echo "ERROR: public key not found at $(PUB_KEY)"; exit 1; }
	@test -f "$(SIGN_KEY)" || { echo "ERROR: private signing key not found at $(SIGN_KEY) -- set SIGN_KEY or write the key there"; exit 1; }
	@echo "==> signing"
	ocm sign component-version \
		--signature $(SIG_NAME) \
		$(CV_REF)

verify: sign
	@echo "==> verifying"
	ocm verify component-version \
		--signature $(SIG_NAME) \
		$(CV_REF)
	@echo "    OK"

bundle-signed: verify
	@echo "==> signed and verified bundle at $(CTF_DIR) ($(COMPONENT):$(VERSION), platforms: $(PLATFORMS))"

# ---------------------------------------------------------------------------
# Publish
# ---------------------------------------------------------------------------

check-publishable:
	@test -n "$(COMPONENT)" || { echo "ERROR: could not read component name from $(CONSTRUCTOR) -- does the file exist?"; exit 1; }
	@_tag=$$(git describe --tags --exact-match HEAD 2>/dev/null) \
		|| { echo "ERROR: refusing to publish $(VERSION): HEAD has no exact tag"; exit 1; }; \
	[ "$${_tag#v}" = "$(VERSION)" ] \
		|| { echo "ERROR: HEAD tag '$$_tag' does not match VERSION '$(VERSION)'"; exit 1; }
	@test -d $(CTF_DIR) || { echo "ERROR: no bundle at $(CTF_DIR) -- run 'make bundle' first"; exit 1; }
	@echo "==> $(COMPONENT):$(VERSION) is publishable"

bundle-airgap: check-publishable
	@echo "==> creating air-gap archive"
	@tar -czf $(AIRGAP_TGZ) -C $(BIN_DIR) ctf
	@cd $(BIN_DIR) && sha256sum $(notdir $(AIRGAP_TGZ)) > $(notdir $(AIRGAP_TGZ)).sha256sum
	@echo "    $(AIRGAP_TGZ)"
	@echo "    $(AIRGAP_TGZ).sha256sum"

publish-registry: check-publishable
	@test -n "$(REGISTRY)" || { echo "ERROR: REGISTRY is not set (e.g. make publish-registry REGISTRY=...)"; exit 1; }
	@echo "==> pushing to $(REGISTRY)"
	ocm transfer component-version \
		$(CV_REF) $(REGISTRY)
	@echo "==> verifying registry copy"
	ocm get component-version \
		$(REGISTRY)//$(COMPONENT):$(VERSION) >/dev/null
	@echo "    published $(REGISTRY)//$(COMPONENT):$(VERSION)"

publish-airgap: bundle-airgap
	@test -n "$(ARTIFACTORY_USER)"  || { echo "ERROR: ARTIFACTORY_USER is not set"; exit 1; }
	@test -n "$(ARTIFACTORY_TOKEN)" || { echo "ERROR: ARTIFACTORY_TOKEN is not set"; exit 1; }
	@test -f "$(AIRGAP_TGZ)"        || { echo "ERROR: airgap archive not found at $(AIRGAP_TGZ)"; exit 1; }
	@echo "==> uploading airgap archive to $(GENERIC_REGISTRY)"
	@_target="$(GENERIC_REGISTRY)/$(NAME)/$(VERSION)/$(notdir $(AIRGAP_TGZ))"; \
	_md5=$$(md5sum "$(AIRGAP_TGZ)" | cut -d' ' -f1); \
	curl --fail --silent --show-error \
		-u "$(ARTIFACTORY_USER):$(ARTIFACTORY_TOKEN)" \
		-H "X-Checksum-MD5: $$_md5" \
		-T "$(AIRGAP_TGZ)" \
		"$(ARTIFACTORY_BASE)/$$_target"
	@echo "==> uploading checksum"
	@_target="$(GENERIC_REGISTRY)/$(NAME)/$(VERSION)/$(notdir $(AIRGAP_TGZ)).sha256sum"; \
	_md5=$$(md5sum "$(AIRGAP_TGZ).sha256sum" | cut -d' ' -f1); \
	curl --fail --silent --show-error \
		-u "$(ARTIFACTORY_USER):$(ARTIFACTORY_TOKEN)" \
		-H "X-Checksum-MD5: $$_md5" \
		-T "$(AIRGAP_TGZ).sha256sum" \
		"$(ARTIFACTORY_BASE)/$$_target"
	@echo "    published $(notdir $(AIRGAP_TGZ)) to $(GENERIC_REGISTRY)/$(NAME)/$(VERSION)/"

publish: publish-registry publish-airgap
