package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"time"
)

func init() {
	// Gobでエンコード/デコードする型を登録
	gob.Register(Memo{})
	gob.Register([]Memo{})
	gob.Register(time.Time{})
}

// EncodeMemosToGob はMemoのスライスをGobエンコードし、Base64文字列として返します。
func EncodeMemosToGob(memos []Memo) (string, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(memos); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

// DecodeGobToMemos はBase64エンコードされたGobデータをMemoのスライスにデコードします。
func DecodeGobToMemos(base64Data string) ([]Memo, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		ConsoleLog("Base64 decode error: " + err.Error())
		return nil, err
	}

	var memos []Memo
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(&memos); err != nil {
		// エラー発生時は空のスライスとエラーを返す
		ConsoleLog("Gob decode error: " + err.Error())
		return nil, err
	}
	return memos, nil
}
