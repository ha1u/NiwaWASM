package main

import (
	"fmt"
	"sort"
	"strings"
	"syscall/js"
	"time"
)

// loadMemosFromLocalStorage はローカルストレージからメモをロードします。
func loadMemosFromLocalStorage() ([]Memo, error) {
	data := js.Global().Get("localStorage").Call("getItem", localStorageKey).String()
	if data == "" || data == "null" { // "null" 文字列もチェック
		return []Memo{}, nil
	}
	memos, err := DecodeGobToMemos(data)
	if err != nil {
		ConsoleLog("Error decoding memos from local storage: " + err.Error() + ". Returning empty list.")
		// データが破損している可能性があるため、空のリストを返し、エラーをログに記録
		return []Memo{}, fmt.Errorf("failed to decode memos: %w", err)
	}
	return memos, nil
}

// saveMemosToLocalStorage はメモをローカルストレージに保存します。
func saveMemosToLocalStorage(memos []Memo) error {
	encodedData, err := EncodeMemosToGob(memos)
	if err != nil {
		return err
	}
	js.Global().Get("localStorage").Call("setItem", localStorageKey, encodedData)
	return nil
}

// addMemo は新しいメモを追加し、更新されたメモリストを返します。
func addMemo(content string) ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		ConsoleLog("Error loading memos for add, potentially corrupted. Attempting to use fresh list: " + err.Error())
		// エラーがあっても、新しい空のリストで続行を試みる
		memos = []Memo{}
	}

	newMemo := Memo{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()), // Unixナノ秒をIDとして使用
		Content:   content,
		CreatedAt: time.Now(),
	}

	memos = append([]Memo{newMemo}, memos...) // 新しいメモを先頭に追加

	sort.SliceStable(memos, func(i, j int) bool {
		return memos[i].CreatedAt.After(memos[j].CreatedAt) // 降順ソート
	})

	if err := saveMemosToLocalStorage(memos); err != nil {
		return nil, err
	}
	return memos, nil
}

// getAllMemos は全てのメモを取得します（降順ソート済み）。
func getAllMemos() ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		ConsoleLog("Error loading all memos: " + err.Error())
		return []Memo{}, err // エラー時は空のスライスとエラーを返す
	}
	sort.SliceStable(memos, func(i, j int) bool {
		return memos[i].CreatedAt.After(memos[j].CreatedAt)
	})
	return memos, nil
}

// deleteMemoByID は指定されたIDのメモを削除します。
func deleteMemoByID(id string) ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		return nil, err
	}

	var updatedMemos []Memo
	found := false
	for _, memo := range memos {
		if memo.ID != id {
			updatedMemos = append(updatedMemos, memo)
		} else {
			found = true
		}
	}

	if !found {
		ConsoleLog("Memo with ID " + id + " not found for deletion.")
		return memos, nil // 見つからなければ元のリストを返す
	}

	if err := saveMemosToLocalStorage(updatedMemos); err != nil {
		return nil, err
	}
	return updatedMemos, nil
}

// searchMemosByKeyword はキーワードでメモを検索します（大文字小文字区別なしの部分一致）。
func searchMemosByKeyword(query string) ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(query) == "" {
		return []Memo{}, nil // クエリが空なら空の結果を返す
	}

	var results []Memo
	lowerQuery := strings.ToLower(query)
	for _, memo := range memos {
		if strings.Contains(strings.ToLower(memo.Content), lowerQuery) {
			results = append(results, memo)
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt) // 結果も降順ソート
	})
	return results, nil
}

// getMemoByID は指定されたIDのメモを取得します。
func getMemoByID(id string) (*Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		return nil, err
	}
	for i := range memos {
		if memos[i].ID == id {
			return &memos[i], nil
		}
	}
	return nil, fmt.Errorf("memo with ID %s not found", id)
}
