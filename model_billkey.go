package main

import (
	"time"
)

const (
	TABLE_BILLKEY = "billkey"
)

type BillkeyModel struct {
	Date       time.Time `db:"billkeydate" json:"date"`
	Data       []byte    `db:"billkey" json:"key"`
	Procedures string    `db:"bkprocs" json:"procedures"`
	Id         int64     `db:"id" json:"id"`
}

func init() {
	dbTables = append(dbTables, DbTable{TableName: TABLE_BILLKEY, Obj: BillkeyModel{}, Key: "Id"})
}
