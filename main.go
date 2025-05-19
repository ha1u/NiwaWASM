package main

import (
	"syscall/js"
)

func registerCallbacks() {
	js.Global().Set("goInitializeApp", js.FuncOf(goInitializeApp))
	js.Global().Set("goSaveMemo", js.FuncOf(goSaveMemo))
	js.Global().Set("goDeleteMemo", js.FuncOf(goDeleteMemo))
	js.Global().Set("goSearchMemos", js.FuncOf(goSearchMemos))
	js.Global().Set("goExportBackupData", js.FuncOf(goExportBackupData)) // 名前変更
	js.Global().Set("goImportData", js.FuncOf(goImportData))
	js.Global().Set("goCopyMemoContent", js.FuncOf(goCopyMemoContent))
	js.Global().Set("goExportFormattedData", js.FuncOf(goExportFormattedData)) // 新規追加
}

func main() {
	c := make(chan struct{}, 0)
	ConsoleLog("庭WASM Go WebAssembly Initialized")
	registerCallbacks()
	<-c
}
