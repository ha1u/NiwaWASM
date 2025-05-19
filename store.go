package main

import (
	"fmt"
	"sort"
	"strings"
	"syscall/js"
	"time"
)

func loadMemosFromLocalStorage() ([]Memo, error) {
	data := js.Global().Get("localStorage").Call("getItem", localStorageKey).String()
	// "null" 文字列や空文字列の場合、DecodeGobToMemos は []Memo{}, nil を返すことを期待
	if data == "null" { // 明示的に "null" 文字列の場合は空として扱う
		return []Memo{}, nil
	}
	// dataが空文字列の場合もDecodeGobToMemosに渡して処理させる (Base64デコードでエラーになるか、空データとして扱われる)
	// DecodeGobToMemosでデータが空の場合の処理が追加されているので、ここで特別に "" をチェックする必要は薄いが、
	// 念のため、非常に短い無効なBase64文字列を渡さないようにする。
	// しかし、 "" は有効な入力として DecodeString に渡せる。
	return DecodeGobToMemos(data)
}

func saveMemosToLocalStorage(memos []Memo) error {
	encodedData, err := EncodeMemosToGob(memos)
	if err != nil {
		return err
	}
	js.Global().Get("localStorage").Call("setItem", localStorageKey, encodedData)
	return nil
}

func addMemo(content string) ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		// このエラーは深刻な破損の場合のみ発生するはず (EOF以外)
		ConsoleLog("Error loading memos for add, potentially corrupted. Attempting to use fresh list: " + err.Error())
		memos = []Memo{}
	}

	newMemo := Memo{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Content:   content,
		CreatedAt: time.Now(),
	}
	memos = append([]Memo{newMemo}, memos...)
	sort.SliceStable(memos, func(i, j int) bool {
		return memos[i].CreatedAt.After(memos[j].CreatedAt)
	})

	if err := saveMemosToLocalStorage(memos); err != nil {
		return nil, err
	}
	return memos, nil
}

func getAllMemos() ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		// 通常、localStorageが空か破損していない限り、深刻なエラーのみがここに到達する。
		ConsoleLog("Error loading all memos in getAllMemos: " + err.Error())
		return []Memo{}, err
	}
	sort.SliceStable(memos, func(i, j int) bool {
		return memos[i].CreatedAt.After(memos[j].CreatedAt)
	})
	return memos, nil
}

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
		return memos, nil
	}
	if err := saveMemosToLocalStorage(updatedMemos); err != nil {
		return nil, err
	}
	return updatedMemos, nil
}

func searchMemosByKeyword(query string) ([]Memo, error) {
	memos, err := loadMemosFromLocalStorage()
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(query) == "" {
		return []Memo{}, nil
	}
	var results []Memo
	lowerQuery := strings.ToLower(query)
	for _, memo := range memos {
		if strings.Contains(strings.ToLower(memo.Content), lowerQuery) {
			results = append(results, memo)
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})
	return results, nil
}

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
