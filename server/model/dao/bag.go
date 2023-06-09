/* *
 * error code: 30001000 ` 30001999
 */

package dao

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/spf13/cast"
	"server/core/xerror"
)

//BagTable Struct
type BagTable struct {
	Uid    int    `json:"uid"`
	Item   string `json:"item"`
	Expire int    `json:"expire"`
	Itime  int    `json:"itime"`
}

//Bag struct
type Bag struct {
	ctx      context.Context
	db       *DBBase
	redisKey string //redis key
	redis    *RedisPoolConn

	tbl    string //表名
	fields []string
}

func NewBag(ctx context.Context) *Bag {
	return &Bag{
		ctx:   ctx,
		db:    NewDBBase(ctx),
		redis: NewRedis(ctx),

		tbl:    "bag",
		fields: []string{"uid", "item", "expire", "itime"},
	}
}

func (d *Bag) genTable(fields []string, values []any) *BagTable {
	entity := BagTable{}
	for k, v := range fields {
		if v == "uid" {
			entity.Uid = cast.ToInt(b2String(*(values[k]).(*sql.RawBytes)))
		}
		if v == "item" {
			entity.Item = b2String(*(values[k]).(*sql.RawBytes))
		}
		if v == "expire" {
			entity.Expire = cast.ToInt(b2String(*(values[k]).(*sql.RawBytes)))
		}
		if v == "itime" {
			entity.Itime = cast.ToInt(b2String(*(values[k]).(*sql.RawBytes)))
		}
	}

	//return
	return &entity
}

func (d *Bag) Query(serverId, uid int, fields []string, where string, order ...string) ([]*BagTable, xerror.Error) {
	if fields == nil {
		fields = d.fields
	}
	d.db.Table(getTableName(uid, d.tbl)).Fields(fields...).Where(where)
	if len(order) > 0 {
		d.db.OrderBy(order[0])
	}

	rows, err := d.db.Query()
	if err != nil {
		return nil, xerror.Wrap(d.ctx, nil, &xerror.TempError{
			Code:    30001000,
			Err:     err,
			Message: "bag.Query",
		})
	}
	defer rows.Close()

	//字段 - 数据
	data := make([]*BagTable, 0, 10)
	entity := make([]any, 0, len(fields))
	for i := 0; i < len(fields); i++ {
		entity = append(entity, new(sql.RawBytes))
	}
	for rows.Next() {
		if err := rows.Scan(entity...); err != nil {
			return nil, xerror.Wrap(d.ctx, nil, &xerror.TempError{
				Code:    30001009,
				Err:     err,
				Message: "bag.Query(dao)",
			})
		}
		data = append(data, d.genTable(fields, entity))
	}
	if len(data) == 0 {
		return data, xerror.Wrap(d.ctx, nil, &xerror.TempError{
			Code:    30001010,
			Err:     ErrorNoRows,
			Message: "bag.Query(dao)",
		})
	}

	return data, nil
}

func (d *Bag) QueryMap(serverId, uid int, fields []string, where string, order ...string) (map[int]*BagTable, xerror.Error) {
	d.db.Table(getTableName(uid, d.tbl)).Fields(fields...).Where(where)
	if len(order) > 0 {
		d.db.OrderBy(order[0])
	}
	rows, err := d.db.Query()
	if err != nil {
		return nil, xerror.Wrap(d.ctx, nil, &xerror.TempError{
			Code:    30001020,
			Err:     err,
			Message: fmt.Sprintf(`bag.QueryMap uid:%v`, uid),
		})
	}
	defer rows.Close()

	//字段 - 数据
	data := make(map[int]*BagTable, 50)
	entity := make([]any, 0, len(fields))
	for i := 0; i < len(fields); i++ {
		entity = append(entity, new(sql.RawBytes))
	}
	for rows.Next() {
		if err := rows.Scan(entity...); err != nil {
			return nil, xerror.Wrap(d.ctx, nil, &xerror.TempError{
				Code:    30001025,
				Err:     err,
				Message: fmt.Sprintf(`bag.QueryMap uid:%v`, uid),
			})
		}
		record := d.genTable(fields, entity)
		data[record.Uid] = record
	}

	return data, nil
}

func (d *Bag) Insert(serverId, uid int, params map[string]any) (int, xerror.Error) {
	id, err := d.db.Table(getTableName(uid, d.tbl)).Insert(params).Exec()
	if err != nil {
		return 0, xerror.Wrap(d.ctx, nil, &xerror.TempError{
			Code:    30001030,
			Err:     err,
			Message: fmt.Sprintf(`serverId:%v, uid:%v, params:%v`, serverId, uid, params),
		})
	}

	return id, nil
}

func (d *Bag) Modify(serverId, uid int, where string, params map[string]any) (int, xerror.Error) {
	count, err := d.db.Table(getTableName(uid, d.tbl)).Where(where).Modify(params).Exec()
	if err != nil {
		return 0, xerror.Wrap(d.ctx, nil, &xerror.TempError{
			Code:    30001040,
			Err:     err,
			Message: "bag.Modify",
		})
	}

	return count, nil
}

func (d *Bag) Delete(serverId, uid int, where string) (int, xerror.Error) {
	count, err := d.db.Table(getTableName(uid, d.tbl)).Where(where).Delete().Exec()
	if err != nil {
		return 0, xerror.Wrap(d.ctx, nil, &xerror.TempError{
			Code:    30001045,
			Err:     err,
			Message: "bag.Delete",
		})
	}

	return count, nil
}
