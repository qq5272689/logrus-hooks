package main

import (
	"github.com/qq5272689/logrus-hooks/TimedRotatingFileHook"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

func main()  {
	log := logrus.New()
	hook, err := TimedRotatingFileHook.NewTRFileHook("/tmp/logs","test.log","test-error.log","H")
	//defer hook.CloseWrites()
	if err!=nil{
		log.Fatalln(err)
		os.Exit(1)
	}
	log.AddHook(hook)
	for i:=0;i<=2;i++{
		time.Sleep(time.Second*1)
		log.Errorln(i)

	}

}

