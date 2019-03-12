////////////////////////////////////////////////////////////////////////////////
// Author:   Nikita Koryabkin
// Email:    Nikita@Koryabk.in
// Telegram: https://t.me/Apologiz
////////////////////////////////////////////////////////////////////////////////

package mailSender

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/mylockerteam/alog"
)

const pathInfoLog = "%s/var/logs/info.log"
const pathErrorLog = "%s/var/logs/error.log"

var logger struct {
	instance *alog.Log
	once     sync.Once
}

func getLogger() *alog.Log {
	pwd, _ := os.Getwd()
	logger.once.Do(func() {
		logger.instance = alog.Create(&alog.Config{
			TimeFormat: time.RFC3339Nano,
			Loggers: alog.LoggerMap{
				alog.LoggerInfo: getInfoLogger(pwd),
				alog.LoggerErr:  getErrorLogger(pwd),
			},
		})
	})
	return logger.instance
}

func getInfoLogger(pwd string) *alog.Logger {
	return &alog.Logger{
		Channel: make(chan string, 100),
		Strategies: []io.Writer{
			alog.GetFileStrategy(fmt.Sprintf(pathInfoLog, pwd)),
			alog.GetDefaultStrategy(),
		},
	}
}

func getErrorLogger(pwd string) *alog.Logger {
	return &alog.Logger{
		Channel: make(chan string, 100),
		Strategies: []io.Writer{
			alog.GetFileStrategy(fmt.Sprintf(pathErrorLog, pwd)),
			alog.GetDefaultStrategy(),
		},
	}
}
