all: test

deps:
	glide install

test:
	go test github.com/luweimy/goutil/syncq
	go test github.com/luweimy/goutil/syncq2
	go test github.com/luweimy/goutil/workerq

benchmark:
	go test github.com/luweimy/goutil/syncq -bench . -benchmem
	go test github.com/luweimy/goutil/syncq2 -bench . -benchmem

clean:
	@rm -rf bin pkg

.PHONY:  deps test clean
