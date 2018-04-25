package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zwj186/alog"
)

func sysSignalHandle() {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGILL, syscall.SIGABRT, syscall.SIGKILL,
		syscall.SIGSEGV, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGSTOP, syscall.SIGTSTP)

	sig := <-sigChan
	alog.Infof("Received System Signal: %d Exit!", sig)
	alog.Stored() //写入缓存日志
	os.Exit(1)
}

func main() {
	go sysSignalHandle()
	alog.RegisterAlog("config.yaml")
	alog.SetLogTag("Sample")
	alog.Debug("Debug info...")
	alog.DebugC("Debug console info...")
	alog.Info("Info info...")
	alog.InfoC("Info console info...")
	alog.Warn("Warn info...")
	alog.WarnC("Warn console info...")
	alog.Error("Error info...")
	alog.ErrorC("Error console info...")
	alog.Fatal("Fatal info...")
	alog.FatalC("Fatal console info...")
	time.Sleep(2 * time.Second)
}
