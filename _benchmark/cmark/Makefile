CMARK_BIN=cmark_benchmark
CMARK_RUN=LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:./cmark-master/build/src ./$(CMARK_BIN)
ifeq ($(OS),Windows_NT)
  CMARK_BIN=cmark_benchmark.exe
  CMARK_RUN=bash -c "PATH=./cmark-master/build/src:$${PATH} ./$(CMARK_BIN)"
endif

.PHONY: run

run: $(CMARK_BIN)
	$(CMARK_RUN)
	go run ./goldmark_benchmark.go

./cmark-master/build/src/config.h:
	wget -nc -O cmark.zip https://github.com/commonmark/cmark/archive/master.zip
	unzip cmark.zip
	rm -f cmark.zip
	cd cmark-master && make

$(CMARK_BIN): ./cmark-master/build/src/config.h
	gcc -I./cmark-master/build/src -I./cmark-master/src  cmark_benchmark.c -o $(CMARK_BIN) -L./cmark-master/build/src -lcmark
