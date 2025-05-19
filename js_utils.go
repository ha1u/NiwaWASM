package main

import (
	"syscall/js"
)

// ConsoleLog はブラウザのコンソールにメッセージを送信します。
func ConsoleLog(message string) {
	js.Global().Get("console").Call("log", message)
}

// Alert はブラウザにアラートダイアログを表示します。
func Alert(message string) {
	js.Global().Call("alert", message)
}

// Confirm はブラウザに確認ダイアログを表示します。
func Confirm(message string) bool {
	return js.Global().Call("confirm", message).Bool()
}

// TriggerFileDownload はブラウザでファイルのダウンロードをトリガーします。
// JavaScript側のヘルパー関数 jsTriggerFileDownload を呼び出します。
func TriggerFileDownload(filename string, mimeType string, base64Data string) {
	js.Global().Call("jsTriggerFileDownload", filename, mimeType, base64Data)
}

// JSCall はグローバルなJavaScript関数を呼び出します。
func JSCall(funcName string, args ...interface{}) js.Value {
	return js.Global().Call(funcName, args...)
}
