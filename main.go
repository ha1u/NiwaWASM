package main

import (
	"syscall/js"
)

// registerCallbacks はGoの関数をJavaScriptに公開します。
func registerCallbacks() {
	js.Global().Set("goInitializeApp", js.FuncOf(goInitializeApp))
	js.Global().Set("goSaveMemo", js.FuncOf(goSaveMemo))
	js.Global().Set("goDeleteMemo", js.FuncOf(goDeleteMemo))
	js.Global().Set("goSearchMemos", js.FuncOf(goSearchMemos))
	js.Global().Set("goExportData", js.FuncOf(goExportData))
	js.Global().Set("goImportData", js.FuncOf(goImportData))
	js.Global().Set("goCopyMemoContent", js.FuncOf(goCopyMemoContent))
}

func main() {
	c := make(chan struct{}, 0) // Goプログラムが終了しないようにチャネルで待機
	ConsoleLog("庭WASM Go WebAssembly Initialized")
	registerCallbacks()
	<-c
}
