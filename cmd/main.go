package main

import (
	. "moussaud.org/petkind/internal"
	. "moussaud.org/petkind/service"
)

func main() {
	LoadConfiguration()
	NewGlobalTracer()
	Start()
}

