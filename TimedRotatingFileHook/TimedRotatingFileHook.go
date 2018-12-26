package TimedRotatingFileHook

import (
	"time"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"errors"
	"path/filepath"
	"fmt"
	"strings"
)

type TRFileHook struct {
	MainWriter *os.File
	ErrWriter *os.File
	MainDateWriter *os.File
	ErrDateWriter *os.File
	mu *sync.Mutex
	NowTime time.Time
	When string
	FileName string
	FileErrName string
	FilePath string
}

type WriteErr struct {
	Errs []error
}

func (e *WriteErr) AddErr(err error)  {
	e.Errs=append(e.Errs,err)
}

func (e WriteErr) Error() string {
	err_strs:=[]string{}
	for i,err:=range e.Errs{
		err_strs=append(err_strs,fmt.Sprintf("WriteErr %d:%s",i,err))
	}
	return strings.Join(err_strs,";")
}


type NewWriterErr struct {
	Errs []error
}

func (e *NewWriterErr) AddErr(err error)  {
	e.Errs=append(e.Errs,err)
}

func (e NewWriterErr) Error() string {
	err_strs:=[]string{}
	for i,err:=range e.Errs{
		err_strs=append(err_strs,fmt.Sprintf("NewWriteErr %d:%s",i,err))
	}
	return strings.Join(err_strs,";")
}


func NewTRFileHook(logdir,filename,when string) (*TRFileHook,error) {
	h := &TRFileHook{}
	h.FilePath=logdir
	h.FileName=filename
	h.FileErrName=filename+"-err"
	h.When=when
	h.NowTime=time.Now()
	if err:=h.newwrite();err!=nil{
		return h,err
	}
	h.mu=new(sync.Mutex)
	return  h,nil
}


func (h *TRFileHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ls,err:=entry.String()
	if err!=nil{
		return err
	}
	lsb:=[]byte(ls)
	nt := time.Now()
	if h.logstr(h.NowTime)!=h.logstr(nt){
		h.updatewrite(nt)
	}
	werr:=new(WriteErr)
	if entry.Level==logrus.ErrorLevel||entry.Level==logrus.FatalLevel||entry.Level==logrus.PanicLevel{
		if _,err:=h.MainDateWriter.Write(lsb);err!=nil{
			werr.AddErr(err)
		}
		if _,err:=h.MainWriter.Write(lsb);err!=nil{
			werr.AddErr(err)
		}
		if _,err:=h.ErrWriter.Write(lsb);err!=nil{
			werr.AddErr(err)
		}
		if _,err:=h.ErrDateWriter.Write(lsb);err!=nil{
			werr.AddErr(err)
		}
	}else {
		if _,err:=h.MainDateWriter.Write(lsb);err!=nil{
			werr.AddErr(err)
		}
		if _,err:=h.MainWriter.Write(lsb);err!=nil{
			werr.AddErr(err)
		}
	}
	if len(werr.Errs)>0{
		return werr
	}
	return nil

}

func (h *TRFileHook) newwrite() (err error) {
	if fi,err:=os.Stat(h.FilePath);err!=nil{
		if err = os.Mkdir(h.FilePath,os.ModePerm);err!=nil{
			return errors.New("目录:"+h.FilePath+"创建失败！！！")
		}
	}else {
		if !fi.IsDir(){
			return errors.New("路径:"+h.FilePath+"非目录！！！")
		}
	}
	nwerr:=new(NewWriterErr)
	timestr := h.logstr(h.NowTime)
	mainfile :=filepath.Join(h.FilePath,h.FileName)
	mainfiledate :=filepath.Join(h.FilePath,h.FileName+"-"+timestr)
	mainfile_flag:=os.O_CREATE|os.O_RDWR|os.O_APPEND
	if fi,err:=os.Stat(mainfile);err==nil&&h.logstr(fi.ModTime())!=h.logstr(h.NowTime){
		mainfile_flag=os.O_CREATE|os.O_RDWR|os.O_TRUNC
	}
	if h.MainWriter,err = os.OpenFile(mainfile,mainfile_flag,0664);err!=nil{
		nwerr.AddErr(err)
	}
	if h.MainDateWriter,err = os.OpenFile(mainfiledate,os.O_CREATE|os.O_APPEND|os.O_RDWR,0664);err!=nil{
		nwerr.AddErr(err)
	}

	errfile :=filepath.Join(h.FilePath,h.FileErrName)
	errfiledate :=filepath.Join(h.FilePath,h.FileErrName+"-"+timestr)

	errfile_flag:=os.O_CREATE|os.O_RDWR|os.O_APPEND
	if fi,err:=os.Stat(errfile);err==nil&&h.logstr(fi.ModTime())!=h.logstr(h.NowTime){
		errfile_flag=os.O_CREATE|os.O_RDWR|os.O_TRUNC
	}

	if h.ErrWriter,err = os.OpenFile(errfile,errfile_flag,0664);err!=nil{
		nwerr.AddErr(err)
	}
	if h.ErrDateWriter,err = os.OpenFile(errfiledate,os.O_CREATE|os.O_APPEND|os.O_RDWR,0664);err!=nil{
		nwerr.AddErr(err)
	}
	if len(nwerr.Errs)>0{
		return nwerr
	}
	return nil

}

func (h *TRFileHook) updatewrite(t time.Time) (err error) {
	h.NowTime = t
	h.CloseWrites()
	return h.newwrite()
}

func (h *TRFileHook) logstr(t time.Time) string {
	switch h.When {
	case "H":
		return t.Format("2006-01-02-15")
	case "M":
		return t.Format("2006-01-02-15-04")
	case "D":
		return t.Format("2006-01-02")
	default:
		return t.Format("2006-01-02-15")
	}
}

func (h *TRFileHook) CloseWrites()  {
	h.MainWriter.Close()
	h.ErrDateWriter.Close()
	h.ErrWriter.Close()
	h.MainDateWriter.Close()
}

func (h *TRFileHook) Levels() []logrus.Level {
	return logrus.AllLevels
}