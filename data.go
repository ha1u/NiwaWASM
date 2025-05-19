package main

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"io" // io.EOF を利用するためにインポート
	"time"
)

func init() {
	gob.Register(Memo{})
	gob.Register([]Memo{})
	gob.Register(time.Time{})
}

func EncodeMemosToGob(memos []Memo) (string, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(memos); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func DecodeGobToMemos(base64Data string) ([]Memo, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		ConsoleLog("Base64 decode error: " + err.Error())
		return nil, err // Base64デコードエラーはそのまま返す
	}

	if len(data) == 0 { // Base64デコード後のデータが空なら、空のメモとして扱う
		return []Memo{}, nil
	}

	var memos []Memo
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	if err := decoder.Decode(&memos); err != nil {
		if err == io.EOF { // EOFエラーは、データが空または正常に終了したとみなし、空のスライスを返す
			ConsoleLog("Gob decode: EOF reached, likely empty data. Returning empty memo list.")
			return []Memo{}, nil
		}
		ConsoleLog("Gob decode error: " + err.Error())
		return nil, err // その他のGobデコードエラー
	}
	return memos, nil
}
