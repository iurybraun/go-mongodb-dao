package logger

import (
	"fmt"
	"time"
	"runtime"
	"os"
	"strings"
)

type CanFormat interface {
	Format() string
}

type HasTrace interface {
	Trace() string
}

func Failed(v ... interface{}) {
	baseErrInfo(v...)
	os.Exit(-1)
}

func Erred(v ... interface{}) {
	baseErrInfo(v...)
}

func Trace(obj interface{}) {
	var info []string
	if ht, ok := obj.(HasTrace); ok {
		s := ht.Trace()
		ss := strings.Split(s, "\n")
		for _, value := range ss {
			info = append(info, fmt.Sprintf("\x1b[0;93m%s [TRACE] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), value))
		}
	} else {
		info = append(info, fmt.Sprintf("\x1b[0;93m%s [TRACE] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), "没有追踪信息"))
	}
	fmt.Print(strings.Join(info, ""))
}

func baseErrInfo(v ... interface{}) {
	var info []string
	for i := 0; i < len(v); i++ {
		if cf, ok := v[i].(CanFormat); ok {
			s := cf.Format()
			ss := strings.Split(s, "\n")
			for _, value := range ss {
				info = append(info, fmt.Sprintf("\x1b[0;31m%s [ERROR] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), value))
			}
		} else if s, ok := v[i].(string); ok {
			ss := strings.Split(s, "\n")
			for _, value := range ss {
				info = append(info, fmt.Sprintf("\x1b[0;31m%s [ERROR] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), value))
			}
		} else {
			info = append(info, fmt.Sprintf("\x1b[0;31m%s [ERROR] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), v[i]))
		}
		if ht, ok := v[i].(HasTrace); ok {
			s := ht.Trace()
			ss := strings.Split(s, "\n")
			for _, value := range ss {
				info = append(info, fmt.Sprintf("\x1b[0;93m%s [TRACE] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), value))
			}
		}
	}
	funcName, file, line, ok := runtime.Caller(2)
	if ok {
		info = append(info, fmt.Sprintf("\x1b[0;93m%s [TRACE] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), runtime.FuncForPC(funcName).Name()))
		info = append(info, fmt.Sprintf("\x1b[0;93m%s [TRACE] %s,line: %d\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), file, line))
	}
	fmt.Print(strings.Join(info, ""))
}

func Info(v ... string) {
	var info []string
	for i := 0; i < len(v); i++ {
		ss := strings.Split(v[i], "\n")
		for _, value := range ss {
			info = append(info, fmt.Sprintf("\x1b[0;34m%s [INFO ] %s\n\x1b[0m", time.Now().Format("2006/01/02 15:04:05"), value))
		}
	}
	fmt.Print(strings.Join(info, ""))
}
