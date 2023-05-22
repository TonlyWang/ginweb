/* *
 * bag
 * error code: 20005000 ` 20005999
 */

package model

import (
	"encoding/json"
	"fmt"
	"server/core/merror"
	"server/core/request"
	"server/global"
	"server/model/dao"
)

//Bag Model
type Bag struct {
	req  *request.Request
	Data map[int]*dao.BagTable
	dao  *dao.Bag
}

func NewBag(req *request.Request) *Bag {
	return &Bag{
		req:  req,
		Data: make(map[int]*dao.BagTable, global.MapInitCount),
		dao:  dao.NewBag(req),
	}
}

func (m *Bag) Get(uid int) merror.Error {
	data, err := m.dao.Get(uid)
	if err.GetError() == dao.ErrorNoRows {
		return merror.NewError(m.req, &merror.ErrorTemp{
			Err:     ErrorNoRows,
			Code:    20230410,
			Message: fmt.Sprintf(`bag.get error:%v`, err.GetError()),
			Type:    1,
		})
	} else if err != nil {
		return merror.NewError(m.req, &merror.ErrorTemp{
			Err:  err.GetError(),
			Code: 20230411,
			Type: 1,
		})
	}
	m.Data = data

	return nil
}

func (m *Bag) Modify(uid int) merror.Error {
	data, err := json.Marshal(m.Data[uid])
	if err != nil {
		return merror.NewError(m.req, &merror.ErrorTemp{
			Err:  err,
			Code: 20230420,
			Type: 1,
		})
	}

	return m.dao.Modify(uid, data)
}

func (m *Bag) Add(uid int) merror.Error {
	data, err := json.Marshal(m.Data[uid])
	if err != nil {
		return merror.NewError(m.req, &merror.ErrorTemp{
			Err:  err,
			Code: 20230430,
			Type: 1,
		})
	}
	if err := m.dao.Modify(uid, data); err.GetError() == dao.ErrorNoRows {
		return merror.NewError(m.req, &merror.ErrorTemp{
			Err:  ErrorNoRows,
			Code: 20230433,
			Type: 1,
		})
	} else if err != nil {
		return merror.NewError(m.req, &merror.ErrorTemp{
			Err:  err.GetError(),
			Code: 20230434,
			Type: 1,
		})
	}

	return nil
}
