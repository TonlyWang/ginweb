package xerror

import (
	"context"
	"fmt"
	"server/core/logger"
)

type TempError struct {
	Err       error
	Code      int32
	Message   string
	ErrorList []Error
	Type      int8 //1 error|2 debug
}

func (e *TempError) GetErr() error {
	return e.Err
}

func (e *TempError) SetErr(err error) {
	e.Err = err
}

func (e *TempError) GetCode() int32 {
	return e.Code
}

func (e *TempError) SetCode(code int32) {
	e.Code = code
}

func (e *TempError) GetMsg() string {
	return e.Message
}

func (e *TempError) SetMsg(msg string) {
	e.Message = msg
}

func (e *TempError) GetType() int8 {
	return e.Type
}

func (e *TempError) SetType(itype int8) {
	e.Type = itype
}

func (e *TempError) Error() string {
	return e.Err.Error()
}

func (e *TempError) AddError(err Error) Error {
	if cap(e.ErrorList) == 0 {
		e.ErrorList = make([]Error, 0, 10)
	}
	e.ErrorList = append(e.ErrorList, err)

	//设置Error为当前最新的Error
	e.SetErr(err.GetErr())
	e.SetCode(err.GetCode())
	e.SetMsg(err.GetMsg())
	e.SetType(err.GetType())

	return e
}

func (e *TempError) GetErrorList() []Error {
	return e.ErrorList
}

func (e *TempError) Copy() Error {
	return &TempError{
		Err:     e.GetErr(),
		Code:    e.GetCode(),
		Message: e.GetMsg(),
	}
}

func (e *TempError) Is(err error) bool {
	if e.GetErr() == err {
		return true
	}
	return false
}

func (e *TempError) Contain(err error) bool {
	for _, v := range e.ErrorList {
		if v.GetErr() == err {
			return true
		}
	}
	return false
}

//Wrap 老的错误信息包裹新的错误信息
//
//@params
//	ctx		context.Context	上下文
//	originalError	  Error	老的Error
//	newError		  Error	新的Error
//@return
//	Error
func Wrap(ctx context.Context, originalError, newError Error) Error {
	if newError == nil {
		panic("the parameter newError cannot be nil")
	}

	switch newError.GetErr().(type) {
	case Error:
		panic(fmt.Sprintf(`the Err field cannot be *Error, error:%v, code:%v, message:%v`, newError, newError.GetCode(), newError.GetMsg()))
	}

	//error
	var err Error
	if originalError == nil {
		err = newError.AddError(newError.Copy())
	} else {
		err = originalError.AddError(newError)
	}

	//log
	if err.GetType() == 1 {
		for _, e := range err.GetErrorList() {
			//fmt.Println("error-list:", e.GetCode(), e.GetErr(), e.GetMsg(), e.GetType())
			//xlog.Errorf(`[%d] message:%v, error:%v`, e.GetCode(), e.GetMsg(), e.GetErr())
			logger.Error(ctx, fmt.Sprintf(`[%d] message:%v, error:%v`, e.GetCode(), e.GetMsg(), e.GetErr()))
		}
	}

	//return
	return err
}
