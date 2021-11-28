BIN=main.out
 
all: deps build run

deps:
	go get github.com/hajimehoshi/ebiten/v2
	go get github.com/hajimehoshi/ebiten/v2/ebitenutil
	go get github.com/nfnt/resize
 
build:
	go build -o $(BIN) .
 
run: build
	./$(BIN) $(filter-out $@, $(MAKECMDGOALS))

%:
	@true
 
clean:
	go clean
	rm $(BIN)