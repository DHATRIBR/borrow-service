package db

import (
    "database/sql"
    "fmt"
    "os"
    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() {
    var err error
    mysqlDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        os.Getenv("MYSQL_USER"),
        os.Getenv("MYSQL_PASSWORD"),
        os.Getenv("MYSQL_HOST"),
        os.Getenv("MYSQL_PORT"),
        os.Getenv("MYSQL_DB"),
    )
    DB, err = sql.Open("mysql", mysqlDSN)
    if err != nil {
        panic(fmt.Sprintf("Failed to connect to DB: %s", err))
    }
    DB.SetMaxOpenConns(10)
    DB.SetMaxIdleConns(5)
}
