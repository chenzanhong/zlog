// test/main.go
package main

import "github.com/chenzanhong/zlog"

func main(){
	zlog.InitDefault()
	zlog.Info("ok")
	zlog.Infof("hello %s", "czh")
	zlog.Panicw("hello", "a","b")
	zlog.Infof("hello %s", "czh")
}