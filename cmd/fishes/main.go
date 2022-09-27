package main

import (
	"moussaud.org/fishes/service/petkind"

	. "moussaud.org/petkind/internal"
)

func main() {
	LoadConfiguration()
	NewGlobalTracer()
	petkind.Start()
}
