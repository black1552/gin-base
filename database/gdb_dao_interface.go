package database

type IDao interface {
	DB() DB
	TableName() string
	Columns() any
}
