package main

import (
	"fmt"
	"strings"
	"syscall/js"
	"time"
)

// memosToJsArray はGoのMemoスライスをJavaScriptのオブジェクト配列に変換します。
func memosToJsArray(memos []Memo) js.Value {
	jsArray := js.Global().Get("Array").New(len(memos))
	for i, m := range memos {
		obj := js.Global().Get("Object").New()
		obj.Set("id", m.ID)
		obj.Set("content", m.Content)
		obj.Set("createdAt", m.CreatedAt.Format(time.RFC3339)) // JSでパースしやすいRFC3339形式
		jsArray.SetIndex(i, obj)
	}
	return jsArray
}

// goInitializeApp はアプリケーション初期化時に呼ばれます。
func goInitializeApp(this js.Value, args []js.Value) interface{} {
	ConsoleLog("Go: Initializing app...")
	memos, err := getAllMemos()
	if err != nil {
		// getAllMemos 内でエラーがログ記録されるので、ここではUIに一般的なエラーを表示
		Alert("記録の読み込み中にエラーが発生しました。")
		JSCall("jsDisplayMemos", memosToJsArray([]Memo{}))
		JSCall("jsUpdateRecordListTitleVisibility", false)
		return nil
	}
	JSCall("jsDisplayMemos", memosToJsArray(memos))
	JSCall("jsUpdateRecordListTitleVisibility", len(memos) > 0)
	ConsoleLog(fmt.Sprintf("Go: Loaded %d memos.", len(memos)))
	return nil
}

// goSaveMemo は新しいメモを保存します。
func goSaveMemo(this js.Value, args []js.Value) interface{} {
	content := args[0].String()
	trimmedContent := strings.TrimSpace(content)

	if trimmedContent == "" {
		Alert("記録内容が空です。")
		return nil
	}

	updatedMemos, err := addMemo(trimmedContent)
	if err != nil {
		Alert("記録の保存に失敗しました: " + err.Error())
		return nil
	}
	JSCall("jsDisplayMemos", memosToJsArray(updatedMemos))
	JSCall("jsClearMemoInput")
	JSCall("jsUpdateRecordListTitleVisibility", len(updatedMemos) > 0)
	return nil
}

// goDeleteMemo は指定されたIDのメモを削除します。
func goDeleteMemo(this js.Value, args []js.Value) interface{} {
	memoID := args[0].String()
	// 確認ダイアログはJavaScript側で表示済みとします。

	updatedMemos, err := deleteMemoByID(memoID)
	if err != nil {
		Alert("記録の削除に失敗しました: " + err.Error())
		return nil
	}
	JSCall("jsDisplayMemos", memosToJsArray(updatedMemos))
	JSCall("jsUpdateRecordListTitleVisibility", len(updatedMemos) > 0)
	Alert("記録を削除しました。")
	return nil
}

// goSearchMemos はキーワードでメモを検索します。
func goSearchMemos(this js.Value, args []js.Value) interface{} {
	query := args[0].String()
	trimmedQuery := strings.TrimSpace(query)

	if trimmedQuery == "" {
		JSCall("jsDisplaySearchResults", memosToJsArray([]Memo{}), "検索キーワードを入力してください。")
		return nil
	}

	results, err := searchMemosByKeyword(trimmedQuery)
	if err != nil {
		Alert("検索処理中にエラーが発生しました: " + err.Error())
		JSCall("jsDisplaySearchResults", memosToJsArray([]Memo{}), "検索エラーが発生しました。")
		return nil
	}

	if len(results) == 0 {
		JSCall("jsDisplaySearchResults", memosToJsArray(results), "検索結果が見つかりませんでした。")
	} else {
		JSCall("jsDisplaySearchResults", memosToJsArray(results), "") // メッセージが空なら結果を表示
	}
	return nil
}

// goExportData は全記録をGob形式でエクスポートします。
func goExportData(this js.Value, args []js.Value) interface{} {
	memos, err := getAllMemos()
	if err != nil {
		Alert("データのエクスポート準備中にエラーが発生しました: " + err.Error())
		return nil
	}

	if len(memos) == 0 {
		Alert("エクスポートする記録がありません。")
		return nil
	}

	gobData, err := EncodeMemosToGob(memos)
	if err != nil {
		Alert("データのエクスポートに失敗しました (エンコードエラー): " + err.Error())
		return nil
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("NiwaWASM_backup_%s.db", timestamp)

	TriggerFileDownload(filename, "application/octet-stream", gobData)
	// Alert("全記録をエクスポートしました。") // jsTriggerFileDownload側でフィードバックがあるかもしれないので省略も可
	return nil
}

// goImportData はGob形式のファイルから記録をインポートします。
func goImportData(this js.Value, args []js.Value) interface{} {
	base64Data := args[0].String()

	if base64Data == "" {
		Alert("インポートするファイルデータが空です。")
		return nil
	}

	importedMemos, err := DecodeGobToMemos(base64Data)
	if err != nil {
		Alert("データのインポートに失敗しました (デコードエラー): " + err.Error() + "\nファイルが破損しているか、形式が正しくない可能性があります。")
		return nil
	}

	if err := saveMemosToLocalStorage(importedMemos); err != nil {
		Alert("インポートしたデータの保存に失敗しました: " + err.Error())
		return nil
	}

	Alert(fmt.Sprintf("%d 件の記録をインポートしました。既存の記録は置き換えられました。", len(importedMemos)))
	JSCall("jsDisplayMemos", memosToJsArray(importedMemos))
	JSCall("jsUpdateRecordListTitleVisibility", len(importedMemos) > 0)
	JSCall("jsShowView", "record") // 記録ビューに切り替えて表示を更新
	return nil
}

// goCopyMemoContent は指定されたメモの内容をクリップボードにコピーするようJSに指示します。
func goCopyMemoContent(this js.Value, args []js.Value) interface{} {
	memoID := args[0].String()
	memo, err := getMemoByID(memoID)
	if err != nil {
		Alert("コピー対象の記録が見つかりません: " + err.Error())
		// JS側で失敗フィードバックを出す想定だが、念のため
		// JSCall("jsCopyFeedback", memoID, false, "コピー失敗")
		return nil
	}
	// JavaScript側の関数 jsCopyToClipboardAndFeedback を呼び出す
	JSCall("jsCopyToClipboardAndFeedback", memo.Content, memoID)
	return nil
}
