package main

import (
	"bytes" // For Markdown/Text export buffer
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort" // For sorting memos for export if needed again
	"strings"
	"syscall/js"
	"text/template" // For simple templating in export
	"time"
)

func memosToJsArray(memos []Memo) js.Value {
	jsArray := js.Global().Get("Array").New(len(memos))
	for i, m := range memos {
		obj := js.Global().Get("Object").New()
		obj.Set("id", m.ID)
		obj.Set("content", m.Content)
		obj.Set("createdAt", m.CreatedAt.Format(time.RFC3339))
		jsArray.SetIndex(i, obj)
	}
	return jsArray
}

func goInitializeApp(this js.Value, args []js.Value) interface{} {
	ConsoleLog("Go: Initializing app...")
	memos, err := getAllMemos()
	if err != nil {
		// ここでのエラーは、Gobデコードに失敗した深刻なケース（EOF以外）を想定。
		// ローカルストレージが空のケースは err == nil で memos が空スライスになる。
		Alert("記録の読み込み中に予期せぬエラーが発生しました: " + err.Error())
		JSCall("jsDisplayMemos", memosToJsArray([]Memo{}))
		JSCall("jsUpdateRecordListTitleVisibility", false)
		return nil
	}
	JSCall("jsDisplayMemos", memosToJsArray(memos))
	JSCall("jsUpdateRecordListTitleVisibility", len(memos) > 0)
	ConsoleLog(fmt.Sprintf("Go: Loaded %d memos.", len(memos)))
	return nil
}

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

func goDeleteMemo(this js.Value, args []js.Value) interface{} {
	memoID := args[0].String()
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

func goSearchMemos(this js.Value, args []js.Value) interface{} {
	query := args[0].String()
	trimmedQuery := strings.TrimSpace(query)

	// JS側で currentSearchResults を保持するため、検索結果をJSONで返す
	results, err := searchMemosByKeyword(trimmedQuery)
	if err != nil {
		Alert("検索処理中にエラーが発生しました: " + err.Error())
		JSCall("jsDisplaySearchResults", memosToJsArray([]Memo{}), "検索エラーが発生しました。", "[]")
		return nil
	}

	// JS側でメッセージを組み立てる
	message := ""
	if trimmedQuery == "" {
		message = "検索キーワードを入力してください。"
	} else if len(results) == 0 {
		message = "検索結果が見つかりませんでした。"
	}

	// 検索結果のメモデータをJSON文字列としてJSに渡す
	resultsJSON, jsonErr := json.Marshal(results)
	if jsonErr != nil {
		Alert("検索結果の処理中にエラーが発生しました: " + jsonErr.Error())
		JSCall("jsDisplaySearchResults", memosToJsArray([]Memo{}), "結果処理エラー。", "[]")
		return nil
	}

	JSCall("jsDisplaySearchResults", memosToJsArray(results), message, string(resultsJSON))
	return nil
}

// goExportBackupData は全記録をGob形式でエクスポートします (旧goExportData)
func goExportBackupData(this js.Value, args []js.Value) interface{} {
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
	filename := fmt.Sprintf("NiwaWASM_backup_%s.data", timestamp)
	TriggerFileDownload(filename, "application/octet-stream", gobData)
	return nil
}

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
	JSCall("jsShowView", "record")
	return nil
}

func goCopyMemoContent(this js.Value, args []js.Value) interface{} {
	memoID := args[0].String()
	memo, err := getMemoByID(memoID)
	if err != nil {
		Alert("コピー対象の記録が見つかりません: " + err.Error())
		return nil
	}
	JSCall("jsCopyToClipboardAndFeedback", memo.Content, memoID)
	return nil
}

// --- 新しいエクスポート関連の関数 ---

const txtExportTemplate = `{{range .}}記録日時: {{.CreatedAt.Format "2006/01/02 15:04:05"}}
内容:
{{.Content}}

{{end}}`

const mdExportTemplate = `# 庭WASM 記録エクスポート

{{range .Memos}}## {{.CreatedAt.Format "2006/01/02 15:04:05"}}
{{.Content}}
---
{{end}}`

func formatMemos(memos []Memo, formatType string) (string, string, string, error) {
	// 時系列が逆順（新しいものが上）になっている可能性があるので、エクスポート時は古いものから順に並び替える
	sort.SliceStable(memos, func(i, j int) bool {
		return memos[i].CreatedAt.Before(memos[j].CreatedAt)
	})

	var buffer bytes.Buffer
	var filenameSuffix, mimeType string
	timestamp := time.Now().Format("20060102_150405")

	switch formatType {
	case "txt":
		filenameSuffix = ".txt"
		mimeType = "text/plain;charset=utf-8"
		tmpl, err := template.New("txtExport").Parse(txtExportTemplate)
		if err != nil {
			return "", "", "", fmt.Errorf("txt template parse error: %w", err)
		}
		if err := tmpl.Execute(&buffer, memos); err != nil {
			return "", "", "", fmt.Errorf("txt template execute error: %w", err)
		}
	case "md":
		filenameSuffix = ".md"
		mimeType = "text/markdown;charset=utf-f-8" // text/markdown or application/octet-stream
		tmpl, err := template.New("mdExport").Parse(mdExportTemplate)
		if err != nil {
			return "", "", "", fmt.Errorf("md template parse error: %w", err)
		}
		data := struct {
			Timestamp string
			Memos     []Memo
		}{
			Timestamp: time.Now().Format("2006/01/02"),
			Memos:     memos,
		}
		if err := tmpl.Execute(&buffer, data); err != nil {
			return "", "", "", fmt.Errorf("md template execute error: %w", err)
		}
	default:
		return "", "", "", fmt.Errorf("unsupported format type: %s", formatType)
	}

	filename := fmt.Sprintf("NiwaWASM_export_%s%s", timestamp, filenameSuffix)
	// JavaScript側でBase64エンコードされた文字列を期待するため、エンコードする
	encodedContent := base64.StdEncoding.EncodeToString(buffer.Bytes())
	return filename, encodedContent, mimeType, nil
}

// goExportFormattedData は指定された形式で全記録またはフィルタリングされた記録をエクスポートします。
// args[0]: formatType (string: "txt", "md")
// args[1]: memosForExportJSON (string: エクスポート対象のメモのJSON文字列、全件の場合は空文字列または"all")
func goExportFormattedData(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		Alert("エクスポート関数の引数が不足しています。")
		return nil
	}
	formatType := args[0].String()
	memosForExportJSON := args[1].String()

	var memosToExport []Memo
	var err error

	if memosForExportJSON == "all" || memosForExportJSON == "" {
		memosToExport, err = getAllMemos()
		if err != nil {
			Alert("エクスポート用の全記録取得に失敗しました: " + err.Error())
			return nil
		}
	} else {
		// JavaScriptから渡されたJSON文字列をデシリアライズ
		err = json.Unmarshal([]byte(memosForExportJSON), &memosToExport)
		if err != nil {
			Alert("エクスポート対象の記録データの解析に失敗しました: " + err.Error())
			return nil
		}
	}

	if len(memosToExport) == 0 {
		Alert("エクスポートする記録がありません。")
		return nil
	}

	filename, contentBase64, mimeType, err := formatMemos(memosToExport, formatType)
	if err != nil {
		Alert("データのエクスポートフォーマット処理中にエラーが発生しました: " + err.Error())
		return nil
	}

	TriggerFileDownload(filename, mimeType, contentBase64)
	return nil
}
