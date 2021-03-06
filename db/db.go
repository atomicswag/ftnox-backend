package db

import (
    . "ftnox.com/common"
    . "ftnox.com/config"
    "time"
    "errors"
    "regexp"
    "database/sql"
    "github.com/lib/pq"
)

var (
    ERR_DUPLICATE_ENTRY = errors.New("ERR_DUPLICATE_ENTRY")
    ERR_SERIAL_TX =       errors.New("ERR_SERIAL_TX")
    ERR_OTHER =           errors.New("ERR_OTHER")

    RE_MYSQL_DUPLICATE_KEY =    regexp.MustCompile("Duplicate entry.*for key '(.+)'")
)

func init() {
    if _db != nil { panic(errors.New("db already set up!")) }
    // connect to the database, ping periodically to maintain the connection
    var err error
    _db, err = sql.Open(Config.DbDriver, Config.DbURL)
    if err != nil { panic(err) }
    go ping()

    // migrate the database
    migrateDb()
}

func ping() {
    defer Recover("Db::Ping")
    ticker := time.Tick(time.Minute)
    for {
        <-ticker
        err := GetDB().Ping()
        if err != nil {
            Debug("DB not ok: %v", err)
        } else {
            Debug("DB ok")
        }
    }
}

// Golang's sql returns an opaque error type.
// We translate these into universal error type.
func GetErrorType(err error) error {
    if err == nil { return nil }

    //dbErr, ok := err.(*mysql.MySQLError)
    dbErr, ok := err.(*pq.Error)

    if ok {

        /*
        // MYSQL
        // http://dev.mysql.com/doc/refman/5.5/en/error-messages-server.html#error_er_dup_entry
        switch dbErr.Number {
        case 1062: return ERR_DUPLICATE_ENTRY
        default:   return ERR_OTHER
        }
        */

        // POSTGRESQL
        // https://github.com/lib/pq/blob/master/error.go
        switch dbErr.Code {
        case "23505" :  return ERR_DUPLICATE_ENTRY
        case "40001" :  return ERR_SERIAL_TX
        default:        return ERR_OTHER
        }

    }
    return err
}

func GetErrorMessage(err error) string {
    if err == nil { return "" }
    //dbErr, ok := err.(*mysql.MySQLError)
    dbErr, ok := err.(*pq.Error)
    if ok {
        return dbErr.Message
    }
    panic("Message unknown for err")
}

func GetDuplicateKey(err error) string {
    //dbErr, ok := err.(*mysql.MySQLError)
    dbErr, ok := err.(*pq.Error)
    if ok {
        /*
        // MYSQL
        message := dbErr.Message
        match := RE_MYSQL_DUPLICATE_KEY.FindStringSubmatch(message)
        if match == nil {
            panic("Not a duplicate key error")
        }
        return match[1]
        */

        return dbErr.Constraint
    }
    panic("Message unknown for err")
}
