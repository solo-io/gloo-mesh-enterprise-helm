#----------------------------------------------------------------------------------
# Versioning
#----------------------------------------------------------------------------------

RELEASE := "true"
ifeq ($(TAGGED_VERSION),)
	TAGGED_VERSION := $(shell git describe --tags --dirty --always)
	RELEASE := "false"
endif
VERSION ?= $(shell echo $(TAGGED_VERSION) | cut -c 2-)

.PHONY: print-version
print-version:
ifeq ($(TAGGED_VERSION),)
	exit 1
endif
	echo $(VERSION)

#----------------------------------------------------------------------------------
# Helm
#----------------------------------------------------------------------------------

CHART_DIR := install/helm/service-mesh-hub-enterprise
OUTPUT_ROOT_DIR := _output
OUTPUT_CHART_DIR := $(OUTPUT_ROOT_DIR)/helm/service-mesh-hub-enterprise

.PHONY: set-version
set-version:
	sed -e 's/%version%/'$(VERSION)'/' $(CHART_DIR)/Chart-template.yaml > $(CHART_DIR)/Chart.yaml

.PHONY: package-chart
package-chart:
	helm dependency update $(CHART_DIR)
	helm package --destination $(OUTPUT_CHART_DIR) $(CHART_DIR)

.PHONY: publish-helm
publish-helm: set-version package-chart
	gsutil -m rsync -r gs://service-mesh-hub-enterprise/service-mesh-hub-enterprise $(OUTPUT_CHART_DIR)
	helm repo index $(OUTPUT_CHART_DIR)
	gsutil -m rsync -r $(OUTPUT_CHART_DIR) gs://service-mesh-hub-enterprise/service-mesh-hub-enterprise

.PHONY: clean-helm
clean-helm:
	rm -rf $(CHART_DIR)/charts $(CHART_DIR)/requirements.lock $(CHART_DIR)/Chart.yaml $(OUTPUT_ROOT_DIR)
