# Configuration
APP_NAME ?= Demo go-uikit
APP_ID ?= games.bombastik.uikit
GO_SRC ?= $(abspath .)

# Internal variables
BUILDER_DIR ?= .build/apk-ebiten-builder
BUILDER_REPO ?= https://github.com/erparts/apk-ebiten-builder
ROOT_DIR ?= $(abspath $(BUILDER_DIR))
INCLUDE_PATH ?= $(BUILDER_DIR)/Include.mk
export APP_ID

$(INCLUDE_PATH):
	git clone $(BUILDER_REPO) $(BUILDER_DIR)

include $(INCLUDE_PATH)