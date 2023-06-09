package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"runtime"
	"server/config"
	"strings"
	"time"
)

// SchemaMeta struct
type SchemaMeta struct {
	DBName    string //数据库名
	TableName string //表
	Field     string //字段名
	Type      string //字段类型
	Comment   string //字段备注
}

// DBBase Struct
type DBBase struct {
	ctx context.Context
	db  *sql.DB

	table     string   //表名
	fields    []string //字段
	values    []any    //字段-值
	where     []string //条件
	order     string   //排序
	group     string   //分组
	have      string   //分组条件
	leftJoin  string   //左关联
	rightJoin string   //右关联
	on        string   //on条件
	sql       string   //sql
}

// db connect struct
var dbConn *sql.DB

func init() {
	if err := createConnDB(context.TODO()); err != nil {
		panic("[create db connect error]")
	} else {
		fmt.Println("[database init successfully]")
	}
}

// createConnDB create db connect
func createConnDB(ctx context.Context) error {
	host := config.Config.Mysql.Host
	port := config.Config.Mysql.Port
	user := config.Config.Mysql.Username
	pass := config.Config.Mysql.Password
	dataSource := fmt.Sprintf("%s:%s@tcp(%s:%s)/test?charset=utf8", user, pass, host, port)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
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

	runtime.SetFinalizer(dbConn, CleanMySQL)

	//return
	return nil
}

// NewDBBase 创建db
func NewDBBase(ctx context.Context) *DBBase {
	if dbConn == nil {
		if err := createConnDB(ctx); err != nil {
			return nil
		}
	}

	//return
	return &DBBase{
		db:     dbConn,
		ctx:    ctx,
		fields: make([]string, 0, 20),
		values: make([]any, 0, 20),
		where:  make([]string, 0, 5),
	}
}

// Table 字段
func (d *DBBase) Table(field string) *DBBase {
	d.table = field

	//return
	return d
}

// Fields 字段
func (d *DBBase) Fields(fields ...string) *DBBase {
	d.fields = append(d.fields, fields...)

	//return
	return d
}

// Where 条件
func (d *DBBase) Where(condition string) *DBBase {
	if len(d.where) == 0 {
		d.where = append(d.where, condition)
	} else {
		d.where = append(d.where, " AND ", condition)
	}

	//return
	return d
}

// ORWhere 条件
func (d *DBBase) ORWhere(condition string) *DBBase {
	if len(d.where) == 0 {
		d.where = append(d.where, condition)
	} else {
		d.where = append(d.where, " OR ", condition)
	}

	//return
	return d
}

// GroupBy 分组
func (d *DBBase) GroupBy(group string) *DBBase {
	d.group = group

	//return
	return d
}

// Having 分组条件
func (d *DBBase) Having(having string) *DBBase {
	d.have = having

	//return
	return d
}

// LeftJoin 关联
func (d *DBBase) LeftJoin(join string) *DBBase {
	d.leftJoin = join

	//return
	return d
}

// RightJoin 关联
func (d *DBBase) RightJoin(join string) *DBBase {
	d.rightJoin = join

	//return
	return d
}

// ON 关联
func (d *DBBase) ON(on string) *DBBase {
	d.on = on

	//return
	return d
}

// OrderBy 排序
func (d *DBBase) OrderBy(order string) *DBBase {
	d.order = order

	//return
	return d
}

// Query 查询数据并返回
func (d *DBBase) Query() (*sql.Rows, error) {
	defer func() {
		//rows.Close()
		d.RestSQL()
	}()

	rawsql, err := d.GenRawSQL()
	if err != nil {
		return nil, err
	}
	d.sql = rawsql
	rows, err := d.db.Query(d.sql)
	if err == sql.ErrNoRows { //这里不会被触发,通常会在Scan时触发error
		return nil, ErrorNoRows
	}
	if err != nil {
		return nil, err
	}

	//return
	return rows, nil
}

// GenRawSQL 生成查询SQL
func (d *DBBase) GenRawSQL() (string, error) {
	if d.table == "" {
		return "", errors.New("table cannot be empty")
	}
	if len(d.fields) == 0 {
		return "", errors.New("field cannot be empty")
	}

	rawsql := fmt.Sprintf("SELECT %v FROM %v", strings.Join(d.fields, ","), d.table)
	if len(d.where) > 0 {
		rawsql = fmt.Sprintf(`%v WHERE %v`, rawsql, strings.Join(d.where, ""))
	}
	if d.group != "" {
		rawsql = fmt.Sprintf(`%v GROUP BY %v`, rawsql, d.order)
	}
	if d.have != "" {
		rawsql = fmt.Sprintf(`%v HAVING %v`, rawsql, d.have)
	}
	if d.leftJoin != "" {
		rawsql = fmt.Sprintf(`%v LEFT JOIN %v`, rawsql, d.leftJoin)
	}
	if d.rightJoin != "" {
		rawsql = fmt.Sprintf(`%v RIGHT JOIN %v`, rawsql, d.rightJoin)
	}
	if d.on != "" {
		rawsql = fmt.Sprintf(`%v ON %v`, rawsql, d.on)
	}
	if d.order != "" {
		rawsql = fmt.Sprintf(`%v ORDER BY %v`, rawsql, d.order)
	}

	//d.sql
	d.sql = rawsql

	//return
	return d.sql, nil
}

// Insert 插入数据
func (d *DBBase) Insert(params map[string]any) *DBBase {
	for k, v := range params {
		d.fields = append(d.fields, k)
		d.values = append(d.values, v)
	}
	d.sql = fmt.Sprintf("INSERT INTO %v(%v) VALUES (%v)", d.table, strings.Join(d.fields, ","), strings.Repeat(",?", len(d.fields))[1:])

	//return
	return d
}

// Modify 修改数据
func (d *DBBase) Modify(params map[string]any) *DBBase {
	for k, v := range params {
		d.fields = append(d.fields, fmt.Sprintf(`%v=?`, k))
		d.values = append(d.values, v)
	}
	d.sql = fmt.Sprintf(`UPDATE %v SET %v WHERE %v`, d.table, strings.Join(d.fields, ","), strings.Join(d.where, ""))

	//return
	return d
}

// Delete 删除数据
func (d *DBBase) Delete() *DBBase {
	d.sql = fmt.Sprintf(`DELETE FROM %v WHERE %v`, d.table, strings.Join(d.where, ""))

	//return
	return d
}

// Exec 执行SQL
func (d *DBBase) Exec() (int, error) {
	defer d.RestSQL()

	if d.sql == "" {
		return 0, errors.New("exec: sql is empty")
	}

	stmt, err := d.db.Prepare(d.sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(d.values...)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, ErrorNoRows
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if id > 0 {
		return int(id), nil
	}

	//return
	return int(count), nil
}

// RestSQL 重置SQL(属性)
func (d *DBBase) RestSQL() {
	d.table = ""
	d.fields = make([]string, 0, 20)
	d.values = make([]any, 0, 20)
	d.where = make([]string, 0, 10)
	d.group, d.have = "", ""
	d.order = ""
	d.leftJoin, d.rightJoin, d.on = "", "", ""

	d.sql = ""
}

// GetSQL 获取SQL(属性)
func (d *DBBase) GetSQL() string {
	return d.sql
}

/* ----------------------------------schema meta---------------------------------- */

// GetTableSchemaMeta 获取表结构
func (d *DBBase) GetTableSchemaMeta(tableName string) ([]SchemaMeta, error) {
	//list, _ := db.Query(fmt.Sprintf(`show columns from %s`, tableName))
	list, err := d.db.Query(fmt.Sprintf("SELECT `TABLE_SCHEMA`,`TABLE_NAME`,`COLUMN_NAME`,`DATA_TYPE`,`COLUMN_COMMENT` FROM `COLUMNS` WHERE TABLE_NAME='%v'", tableName))
	if err != nil {
		return nil, err
	}
	defer list.Close()

	metas := make([]SchemaMeta, 0, 50)
	for list.Next() {
		var data SchemaMeta
		err := list.Scan(&data.DBName, &data.TableName, &data.Field, &data.Type, &data.Comment)
		if err != nil {
			return nil, err
		}
		metas = append(metas, data)
	}

	return metas, nil
}

// GenTableStruct 生成表表结构Struct
func (d *DBBase) GenTableStruct(tableName string, metas []SchemaMeta) string {
	var fieldValue string

	//字段处理
	for _, v := range metas {
		ftype := "any"
		if strings.Contains(v.Type, "int") {
			ftype = "int"
		} else if strings.Contains(v.Type, "char") {
			ftype = "string"
		} else if strings.Contains(v.Type, "datetime") {
			ftype = "time.Time"
		}

		field := v.Field
		if strings.Contains(field, "_") {
			fields := strings.Split(field, "_")
			for k, v := range fields {
				fields[k] = fmt.Sprintf(`%s%s`, strings.ToUpper(v[:1]), v[1:])
			}
			field = fmt.Sprintf(`%s`, strings.Join(fields, ""))
		} else {
			field = strings.ToUpper(field[:1]) + field[1:]
		}

		comment := ""
		if v.Comment != "" {
			comment = "//" + v.Comment
		}
		fieldValue += fmt.Sprintf("%s %s	`json:\"%v\"` %s \n", field, ftype, strings.ToLower(v.Field), comment)
	}

	//表名处理
	if strings.Contains(tableName, "_") {
		tblName := strings.Split(tableName, "_")
		for k, v := range tblName {
			tblName[k] = fmt.Sprintf(`%s%s`, strings.ToUpper(v[:1]), v[1:])
		}
		tableName = fmt.Sprintf(`%s%s`, strings.Join(tblName, ""), "Table")
	}

	//备注
	structComment := fmt.Sprintf("//%v Struct \n", tableName)

	//return
	return fmt.Sprintf("%stype %s struct {\n%s}", structComment, tableName, fieldValue)
}

/* ----------------------------------function---------------------------------- */

func CleanMySQL(db *sql.DB) {
	db.Close()
}
