package TimedRotatingFileHook

import (
	"time"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"errors"
	"path/filepath"
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

func NewTRFileHook(logdir,filename,errfilename,when string) (*TRFileHook,error) {
	h := &TRFileHook{}
	h.FilePath=logdir
	h.FileName=filename
	h.FileErrName=errfilename
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
	if h.logstr(h.NowTime)!=h.logstr(entry.Time){
		nt := time.Now()
		if h.NowTime.UnixNano()<entry.Time.UnixNano()&&h.logstr(nt)==h.logstr(entry.Time){
			h.updatewrite(entry.Time)
		}
	}
	if entry.Level==logrus.ErrorLevel||entry.Level==logrus.FatalLevel||entry.Level==logrus.PanicLevel{
		errstr:=""
		if _,err:=h.MainDateWriter.Write(lsb);err!=nil{
			errstr+="MainDateWriter err:"+err.Error()+";"
		}
		if _,err:=h.MainWriter.Write(lsb);err!=nil{
			errstr+="MainWriter err:"+err.Error()+";"
		}
		if _,err:=h.ErrWriter.Write(lsb);err!=nil{
			errstr+="ErrWriter err:"+err.Error()+";"
		}
		if _,err:=h.ErrDateWriter.Write(lsb);err!=nil{
			errstr+="ErrDateWriter err:"+err.Error()+";"
		}
		if len(errstr)>0{
			return errors.New(errstr)
		}


	}else {
		errstr:=""
		if _,err:=h.MainDateWriter.Write(lsb);err!=nil{
			errstr+="MainDateWriter err:"+err.Error()+";"
		}
		if _,err:=h.MainWriter.Write(lsb);err!=nil{
			errstr+="MainWriter err:"+err.Error()+";"
		}
		if len(errstr)>0{
			return errors.New(errstr)
		}
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
	timestr := h.logstr(h.NowTime)
	mainfile :=filepath.Join(h.FilePath,h.FileName)
	mainfiledate :=filepath.Join(h.FilePath,h.FileName+"-"+timestr)
	if h.MainWriter,err = os.OpenFile(mainfile,os.O_CREATE|os.O_RDWR|os.O_TRUNC,0664);err!=nil{
		return errors.New("文件:"+mainfile+"打开失败！！！")
	}
	if h.MainDateWriter,err = os.OpenFile(mainfiledate,os.O_CREATE|os.O_APPEND|os.O_RDWR,0664);err!=nil{
		return errors.New("文件:"+mainfiledate+"打开失败！！！")
	}

	errfile :=filepath.Join(h.FilePath,h.FileErrName)
	errfiledate :=filepath.Join(h.FilePath,h.FileErrName+"-"+timestr)
	if h.ErrWriter,err = os.OpenFile(errfile,os.O_CREATE|os.O_RDWR|os.O_TRUNC,0664);err!=nil{
		return errors.New("文件:"+errfile+"打开失败！！！")
	}
	if h.ErrDateWriter,err = os.OpenFile(errfiledate,os.O_CREATE|os.O_APPEND|os.O_RDWR,0664);err!=nil{
		return errors.New("文件:"+errfiledate+"打开失败！！！")
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