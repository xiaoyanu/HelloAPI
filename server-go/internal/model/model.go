package model

import "encoding/json"

type UserInfo struct {
	ID       int    `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Nick     string `db:"nick" json:"nick"`
	Mode     int    `db:"mode" json:"mode"`
	Mail     string `db:"mail" json:"mail"`
}

type User struct {
	UserInfo
	Password *string       `db:"password" json:"password"`
	Created  LocalDateTime `db:"created" json:"created"`
	Updated  LocalDateTime `db:"updated" json:"updated"`
}

type UserKey struct {
	UserID  int           `db:"user_id" json:"user_id"`
	Key     string        `db:"key" json:"key"`
	Created LocalDateTime `db:"created" json:"created"`
	Updated LocalDateTime `db:"updated" json:"updated"`
}

type APIApp struct {
	ID            int           `db:"id" json:"id"`
	Title         *string       `db:"title" json:"title"`
	SmallTitle    *string       `db:"smallTitle" json:"smallTitle"`
	Status        int           `db:"status" json:"status"`
	Type          int           `db:"type" json:"type"`
	URL           *string       `db:"url" json:"url"`
	SendType      int           `db:"sendType" json:"sendType"`
	ReturnType    *string       `db:"returnType" json:"returnType"`
	ReturnContent *string       `db:"returnContent" json:"returnContent"`
	Created       LocalDateTime `db:"created" json:"created"`
	Updated       LocalDateTime `db:"updated" json:"updated"`
	UserID        int           `db:"user_id" json:"user_id"`
	ViewStatus    int           `db:"view_status" json:"view_status"`
}

type APIParam struct {
	APIID    int     `db:"api_id" json:"api_id"`
	Name     *string `db:"name" json:"name"`
	Required int     `db:"required" json:"required"`
	Type     *string `db:"type" json:"type"`
	Msg      *string `db:"msg" json:"msg"`
}

type APIKey struct {
	APIID   int            `db:"api_id" json:"api_id"`
	Key     string         `db:"key" json:"key"`
	Created LocalDateTime  `db:"created" json:"created"`
	Updated LocalDateTime  `db:"updated" json:"updated"`
	Type    int            `db:"type" json:"type"`
	Started *LocalDateTime `db:"started" json:"started"`
	Expired *LocalDateTime `db:"expired" json:"expired"`
	Count   int64          `db:"count" json:"count"`
	Desc    *string        `db:"desc" json:"desc"`
}

type APIRequestLog struct {
	APIID  int              `db:"api_id" json:"app_id"`
	IP     *string          `db:"ip" json:"ip"`
	Time   LocalDateTime    `db:"time" json:"time"`
	Header *json.RawMessage `db:"header" json:"header"`
	Body   *json.RawMessage `db:"body" json:"body"`
	APIKey *string          `db:"api_key" json:"api_key"`
	UserID int              `db:"user_id" json:"user_id"`
}

type APIView struct {
	APIID int   `db:"api_id" json:"app_id"`
	Count int64 `db:"count" json:"count"`
}

type Link struct {
	ID      int           `db:"link_id" json:"link_id"`
	Name    string        `db:"name" json:"name"`
	URL     string        `db:"url" json:"url"`
	Desc    *string       `db:"desc" json:"desc"`
	Avatar  *string       `db:"avatar" json:"avatar"`
	Created LocalDateTime `db:"created" json:"created"`
	Updated LocalDateTime `db:"updated" json:"updated"`
}

type Setting struct {
	Key     string        `db:"key" json:"key"`
	Value   string        `db:"value" json:"value"`
	Updated LocalDateTime `db:"updated" json:"updated"`
}
