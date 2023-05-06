/*
 * error code: 30001000 ` 30001999
 */

package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"server/config"
	"server/core/logger"
	"time"
)

//DBBase struct
type DBBase struct {
	db  *sql.DB
	ctx context.Context
}

//db connect struct
var dbConn *sql.DB

func init() {
	if err := createConnDB(context.TODO()); err != nil {
		panic("[create db connect error]")
	} else {
		fmt.Println("[database init successfully] host:", config.Config.Mysql.Host)
	}
}

//create db connect
func createConnDB(ctx context.Context) error {
	host := config.Config.Mysql.Host
	port := config.Config.Mysql.Port
	user := config.Config.Mysql.Username
	pass := config.Config.Mysql.Passworld
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/test?charset=utf8", user, pass, host, port)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		logger.Error(ctx, fmt.Sprintf(`mysql connect is error: %v`, err))
		return err
	}
	if err := db.Ping(); err != nil {
		logger.Error(ctx, fmt.Sprintf(`mysql ping is error: %v`, err))
		return err
	}
	//设置连接可以重用的最长时间
	db.SetConnMaxLifetime(5 * time.Minute)
	//设置与数据库的最大打开连接数,如果 n <= 0,则对打开的连接数没有限制。默认值为 0（无限制）。
	db.SetMaxOpenConns(10)
	//设置空闲连接池的最大连接数,如果 n <= 0，则不保留任何空闲连接。
	db.SetMaxIdleConns(10)
	//设置连接可能处于空闲状态的最长时间
	db.SetConnMaxIdleTime(5 * time.Minute)
	//dbConn
	dbConn = db

	//return
	return nil
}

//NewDBBase
func NewDBBase(ctx context.Context) *DBBase {
	if dbConn == nil {
		if err := createConnDB(ctx); err != nil {
			logger.Error(ctx, "[1000100] create db connect err")
			return nil
		}
	}

	//return
	return &DBBase{
		db:  dbConn,
		ctx: ctx,
	}
}

//get
func (d *DBBase) Get(table string, uid int) (any, error) {
	sqlText := fmt.Sprintf(`SELECT content FROM %v WHERE uid = ? LIMIT 1`, getTableName(uid, table))
	stmt, err := d.db.Prepare(sqlText)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000110] get error: %v`, err))
		return nil, err
	}
	defer stmt.Close()

	var data any
	if err := stmt.QueryRow(uid).Scan(&data); err == sql.ErrNoRows {
		defaultData, err := d.initData(table, uid)
		if err != nil {
			logger.Error(d.ctx, fmt.Sprintf(`[1000112] get error: %v`, err))
			return nil, err
		}
		return defaultData, nil
	} else if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000114] get error: %v`, err))
		return nil, err
	}
	if data != nil {
		return data, nil
	}

	logger.Error(d.ctx, "[1000119] get unknow error")
	return nil, errors.New("unknow error")
}

//modify
func (d *DBBase) Modify(table string, uid int, data []byte) error {
	sql := fmt.Sprintf(`UPDATE %v SET content=?, stime=? WHERE uid=? LIMIT 1`, getTableName(uid, table))
	stmt, err := d.db.Prepare(sql)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000120] modify error: %v`, err))
		return err
	}
	defer stmt.Close()

	result, err := stmt.Exec(data, time.Now().Unix(), uid)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000122] modify error: %v`, err))
		return err
	}
	num, err := result.RowsAffected()
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000124] modify error: %v`, err))
		return err
	}
	if num == 1 {
		return nil
	}
	if num == 0 {
		logger.Error(d.ctx, "[1000128] modify error")
		return ErrorNoRows
	}

	logger.Error(d.ctx, `[1000129] modify unknow error`)
	return errors.New("unknow error")
}

//init
func (d *DBBase) initData(table string, uid int) (any, error) {
	defaultData := config.GetDefaultDBValue(table)
	sqltext := fmt.Sprintf(`INSERT INTO %v(uid, content, stime) VALUES(?, ?, ?)`, getTableName(uid, table))
	stmt, err := d.db.Prepare(sqltext)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000130] init data error: %v`, err))
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(uid, defaultData, time.Now().Unix())
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000132] init data error: %v`, err))
		return nil, err
	}
	num, err := result.RowsAffected()
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000134] init data error: %v`, err))
		return nil, err
	}
	if num == 0 {
		logger.Error(d.ctx, fmt.Sprintf(`[1000138] init data error: %v`, err))
		return nil, ErrorNoRows
	} else if num == 1 {
		return defaultData, nil
	}

	logger.Error(d.ctx, "[1000139] nuknow error")
	return nil, errors.New("nuknow error")
}

/* ----------------------------------run raw sql---------------------------------- */

//查询单条数据并返回error
func (d *DBBase) QueryRow(rawsql string, scanArgs ...any) error {
	if err := d.db.QueryRow(rawsql).Scan(scanArgs...); err == sql.ErrNoRows {
		logger.Error(d.ctx, fmt.Sprintf(`[1000140] query row error: %v`, err))
		return ErrorNoRows
	} else if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000142] query row error: %v`, err))
		return err
	}

	//return
	return nil
}

//查询多条数据并返回*sql.Row、error
func (d *DBBase) Query(rawsql string) (*sql.Rows, error) {
	rows, err := d.db.Query(rawsql)
	if err == sql.ErrNoRows { //这里不会被触发,通常会在Scan时触发error
		//logger.Error(d.ctx, fmt.Sprintf(`[1000150] query error: %v`, err))
		return nil, ErrorNoRows
	}
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000154] query error: %v`, err))
		return nil, err
	}
	//defer rows.Close()

	return rows, nil
}

//执行SQL并返回是否成功
func (d *DBBase) Exec(rawsql string, args ...any) error {
	result, err := d.db.Exec(rawsql, args...)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000155] exec error: %v`, err))
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000156] exec error: %v`, err))
		return err
	}
	if n == 0 {
		logger.Error(d.ctx, "[1000158] exec the affected rows is the zero.")
		return ErrorNoRows
	}
	if n > 0 {
		return nil
	}

	logger.Error(d.ctx, "[1000159] exec unknow error.")
	return errors.New("exec unknow error.")
}

//插入并返回error
func (d *DBBase) Inert(rawsql string, args ...any) error {
	result, err := d.db.Exec(rawsql, args...)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000170] insert error: %v`, err))
		return err
	}
	if n, err := result.RowsAffected(); err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000175] insert error: %v`, err))
		return err
	} else if n == 0 {
		logger.Error(d.ctx, fmt.Sprintf(`[1000178] insert error: %v`, "the affected rows is the zero"))
		return ErrorNoRows
	} else if n > 0 {
		return nil
	}

	logger.Error(d.ctx, "[1000169] insert unknow error")
	return errors.New("insert unknow error")
}

//插入并返回ID
func (d *DBBase) InertID(rawsql string, args ...any) (int64, error) {
	result, err := d.db.Exec(rawsql, args...)
	if err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000180] insert id error: %v`, err))
		return 0, err
	}
	if n, err := result.RowsAffected(); err != nil {
		logger.Error(d.ctx, fmt.Sprintf(`[1000182] insert id error: %v`, err))
		return 0, err
	} else if n == 0 {
		logger.Error(d.ctx, fmt.Sprintf(`[1000184] insert id error: %v`, "affected row is the zero"))
		return 0, ErrorNoRows
	} else if n == 1 {
		newId, err := result.LastInsertId()
		if err != nil {
			logger.Error(d.ctx, fmt.Sprintf(`[1000188] insert id error: %v`, err))
			return 0, err
		}
		return newId, nil
	}

	logger.Error(d.ctx, `[1000189] insert id unknow error`)
	return 0, errors.New("insert id unknow error")
}

//程序结束之后的清理工作
func FinishClear() {
	dbConn.Close()
	dbConn = nil
}
