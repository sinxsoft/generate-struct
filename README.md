# generate-struct
generate struct from mysql table for golang, a small userful tool


how to use
cd generate-struct

1、change const variables fro db config
vi generate.go

const (
	DB_TYPE = "mysql"
	DB_HOST = "127.0.0.1"
	DB_PORT = "3306"
	DB_USER = "root"
	DB_PASS = "root"
	DB_NAME = "dbname"
)

2、type this
//table_name 数据库表名
go run generate.go  table_name  


