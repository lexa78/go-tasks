package main

import (
	"io"
	"os"
	"runtime"
	"runtime/pprof"
)

func main() {
	f, _ := os.Create("profileCpuFast.pb.gz")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	err := pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	FastSearch(io.Discard)
	///////////////////////////////////////
	f, _ = os.Create("profileMemFast.pb.gz")
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	// Сначала желательно вызвать сборку мусора для точности
	runtime.GC()

	// Записываем текущий профиль памяти
	if err := pprof.WriteHeapProfile(f); err != nil {
		panic(err)
	}
	///////////////////////////////////////

}
