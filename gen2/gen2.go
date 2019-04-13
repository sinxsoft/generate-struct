package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"gopkg.in/fatih/set.v0"
	"os"
	"os/exec"
	"strings"
)

type FieldInfo struct {
	ColName    string `db:"COLUMN_NAME"`
	DataType   string `db:"DATA_TYPE"`
	ColComment string `db:"COLUMN_COMMENT"`
	IsNullable string `db:"IS_NULLABLE"`
}

var dbname = DB_NAME //flag.String("db", "", "the database name")
//var tblname = flag.String("tbl", "", "the table name to export")
var savepath = flag.String("path", "/Users/henrik/Documents/go-lang-orm-file/", "the path to save file")

const (
	DB_TYPE = "mysql"
	dbhost  = "rm-bp1n3v7z0on4i8po0oo.mysql.rds.aliyuncs.com"
	DB_PORT = "3306"
	dbuser  = "miaosu_admin"
	dbpwd   = "ahgcDkrK51xRKPn2Nj"
	DB_NAME = "qiandingdang"
)

func fmtFieldDefine(src string) string {
	temp := strings.Split(src, "_") // 有下划线的，需要拆分
	var str string
	for i := 0; i < len(temp); i++ {
		b := []rune(temp[i])
		for j := 0; j < len(b); j++ {
			if j == 0 {
				// 首字母大写转换
				b[j] -= 32
				str += string(b[j])
			} else {
				str += string(b[j])
			}
		}
	}

	return str
}

func showTables2() set.Interface {
	db, err := sql.Open("mysql", dbuser+":"+dbpwd+"@tcp("+dbhost+":"+DB_PORT+")/"+DB_NAME)
	if err != nil {
		panic(err)
	}
	query, err := db.Query("select table_name from information_schema.tables where table_schema='qiandingdang' and table_type='base table'")
	a := set.New(set.ThreadSafe)
	tn := ""
	for query.Next() {
		if err := query.Scan(&tn); err != nil { //query.Scan查询出来的不定长值放到scans[i] = &values[i],也就是每行都放在values里
			fmt.Println(err)
			break
		}
		a.Add(tn)

	}
	return a
}

func main() {
	//flag.Parse()

	iff := showTables2()

	list := iff.List()
	for _, tblname := range list {

		fmt.Println("table name -->", tblname)

		dns := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8", dbuser, dbpwd, dbhost, "information_schema")

		db := sqlx.MustConnect("mysql", dns)
		//db, err := sql.Open("mysql", dbuser+":"+dbpwd+"@tcp("+dbhost+":"+DB_PORT+")/"+DB_NAME)
		//if err != nil {
		//	panic(err)
		//}

		var fs []FieldInfo
		err := db.Select(&fs, "SELECT COLUMN_NAME, DATA_TYPE, COLUMN_COMMENT, IS_NULLABLE FROM COLUMNS WHERE TABLE_NAME=? and table_schema=?", tblname.(string), dbname)
		if err != nil {

			fmt.Println(err)
			panic(err)
		}

		if len(fs) > 0 {
			var buffer bytes.Buffer
			buffer.WriteString("package models\n")
			buffer.WriteString("\n")
			buffer.WriteString("import (\n")
			buffer.WriteString("\"database/sql\"\n")
			buffer.WriteString("\"time\"\n")
			buffer.WriteString(")\n")
			buffer.WriteString("type " + fmtFieldDefine(tblname.(string)) + " struct {\n")
			for _, v := range fs {
				buffer.WriteString("" + fmtFieldDefine(v.ColName) + " ")
				switch v.DataType {
				case "int", "tinyint", "smallint":
					if v.IsNullable == "YES" {
						buffer.WriteString("sql.NullInt64 ")
					} else {
						buffer.WriteString("int ")
					}
				case "bigint":
					if v.IsNullable == "YES" {
						buffer.WriteString("sql.NullInt64 ")
					} else {
						buffer.WriteString("int64 ")
					}
				case "char", "varchar", "longtext", "text", "tinytext":
					if v.IsNullable == "YES" {
						buffer.WriteString("sql.NullString ")
					} else {
						buffer.WriteString("string ")
					}
				case "date", "datetime", "timestamp":
					buffer.WriteString("time.Time ")
				case "double", "float":
					if v.IsNullable == "YES" {
						buffer.WriteString("sql.NullFloat64 ")
					} else {
						buffer.WriteString("float64 ")
					}
				default:
					// 其他类型当成string处理
					if v.IsNullable == "YES" {
						buffer.WriteString("sql.NullString ")
					} else {
						buffer.WriteString("string ")
					}
				}

				buffer.WriteString(fmt.Sprintf("`db:\"%s\" json:\"%s\"`\n", v.ColName, v.ColName))

			}
			buffer.WriteString(`}`)

			fmt.Println(buffer.String())

			filename := *savepath + tblname.(string) + ".go"
			f, _ := os.Create(filename)
			f.Write([]byte(buffer.String()))
			f.Close()

			cmd := exec.Command("goimports", "-w", filename)
			cmd.Run()
		} else {
			fmt.Println("查询不到数据")
		}
	}
}
