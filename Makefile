all: test

deps:
	glide install

test:
	go test github.com/luweimy/goutil/loop
	go test github.com/luweimy/goutil/syncq
	go test github.com/luweimy/goutil/workerq

benchmark:
	go test github.com/luweimy/goutil/loop -bench . -benchmem
	go test github.com/luweimy/goutil/syncq -bench . -benchmem

clean:
	@rm -rf bin pkg

.PHONY:  deps test clean
