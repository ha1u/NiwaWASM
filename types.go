package main

import "time"

const localStorageKey = "niwaWASM_memos"

// Memo は一つの記録を表す構造体です。
type Memo struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}
