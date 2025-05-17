WASM_EXEC := $(shell go env GOROOT)/lib/wasm/wasm_exec.js
OUT_DIR=../../docs/playground
GOLDMARK_VERSION = $(shell bash -c "")

all: $(OUT_DIR)/goldmark-playground.wasm $(OUT_DIR)/index.html $(OUT_DIR)/playground.css

.PHONY: run
run:
	python3 -m http.server --directory $(OUT_DIR)

$(OUT_DIR)/index.html: index.html $(OUT_DIR)/goldmark-playground.wasm
	mkdir -p $(@D) 
	version=$$(cat wasm/go.mod | grep goldmark | awk '{print $$3}'); cat $< | sed -e "s!{{version}}!$${version}!g" > $@


$(OUT_DIR)/playground.css: playground.css
	mkdir -p $(@D) 
	cp $< $@

$(OUT_DIR)/goldmark-playground.wasm: wasm/main.go wasm/go.mod $(OUT_DIR)/wasm_exec.js
	mkdir -p $(@D) 
	cd wasm; GOOS=js GOARCH=wasm go build -o ../$@ $$(basename "$<")

$(OUT_DIR)/wasm_exec.js: $(WASM_EXEC)
	mkdir -p $(@D) 
	cp $< $@

.PHONY: clean
clean:
	rm -rf $(OUT_DIR)
