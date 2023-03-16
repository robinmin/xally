package utility

import (
	"fmt"
	"net/http"
	"os/user"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/google/uuid"
	"github.com/robinmin/xally/config"
	log "github.com/sirupsen/logrus"
)

type UserDefinedEvent int
type UserDefinedEventMeta struct {
	Name  string
	Level sentry.Level
	Group string
}

const (
	// execute host command on local machine
	EVT_CLIENT_INIT UserDefinedEvent = iota
	EVT_CLIENT_CLOSE
	EVT_CLIENT_ASK_CHATGPT
	EVT_CLIENT_ANSWER_CHATGPT

	EVT_SERVER_INIT
	EVT_SERVER_CLOSE
	EVT_SERVER_PROXY_SUCCESS
	EVT_SERVER_PROXY_FAILED
)

var sentry_events_meta = map[UserDefinedEvent]UserDefinedEventMeta{
	EVT_CLIENT_INIT:           {Name: "evt_client_init", Level: "info", Group: "sys"},
	EVT_CLIENT_CLOSE:          {Name: "evt_client_close", Level: "info", Group: "sys"},
	EVT_CLIENT_ASK_CHATGPT:    {Name: "evt_client_ask_chatgpt", Level: "info", Group: "sys"},
	EVT_CLIENT_ANSWER_CHATGPT: {Name: "evt_client_answer_chatgpt", Level: "info", Group: "sys"},

	EVT_SERVER_INIT:          {Name: "evt_server_init", Level: "info", Group: "sys"},
	EVT_SERVER_CLOSE:         {Name: "evt_server_close", Level: "info", Group: "sys"},
	EVT_SERVER_PROXY_SUCCESS: {Name: "evt_server_proxy_success", Level: "info", Group: "sys"},
	EVT_SERVER_PROXY_FAILED:  {Name: "evt_server_proxy_failed", Level: "info", Group: "sys"},
}

const sentry_event_level = log.InfoLevel

///////////////////////////////////////////////////////////////////////////////

// InitSentry 初始化sentry
func InitSentry(dsn string, is_client bool) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return fmt.Errorf("Failed to initialize sentry: %v", err)
	}

	if is_client {
		// get user id
		current_user, err := user.Current()
		if err == nil {
			log.Error("Failed to get current user information: %v", err.Error())
		} else {
			sentry.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetUser(sentry.User{
					ID:       current_user.Uid,
					Username: current_user.Username,
					Name:     current_user.Name,
				})
			})
		}
	}
	return nil
}

func SetUser(id string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID: id,
		})
	})
}

func SetTag(key, value string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, value)
	})
}

func SetExtra(key string, value interface{}) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetExtra(key, value)
	})
}

// RecoverPanic 恢复panic并发送到sentry
func CloseSentry() {
	// 兜底所有的异常监控与处理
	if err := recover(); err != nil {
		sentry.CurrentHub().Recover(err)

		// 确保所有事件都被发送到Sentry
		sentry.Flush(time.Second * 5)
	}
}

// CaptureException 捕获异常并发送到sentry
func CaptureException(err error) {
	sentry.CaptureException(err)
}

// CaptureRequest 捕获请求并发送到sentry
func CaptureRequest(r *http.Request) {
	hub := sentry.CurrentHub().Clone()
	hub.Scope().SetRequest(r)
}

// ReportCustomEvent 上报定制事件
func ReportEvent(event_id UserDefinedEvent, eventMessage string, payLoad map[string]interface{}) {
	if config.MyConfig.System.SentryDSN == "" {
		// keep silient if not neeed it
		return
	}

	needReport, meta := getEventConfig(event_id)
	if needReport {
		event := sentry.NewEvent()
		event.Level = meta.Level
		event.Message = eventMessage
		event.Tags = map[string]string{
			"event_name": meta.Name,
			"event_id":   uuid.New().String(),
		}
		if payLoad != nil {
			event.Extra = payLoad
		}
		sentry.CaptureEvent(event)
	}
}

func getEventConfig(event_id UserDefinedEvent) (bool, *UserDefinedEventMeta) {
	var meta UserDefinedEventMeta
	var ok bool
	var needReport bool

	if meta, ok = sentry_events_meta[event_id]; !ok {
		// by default report all
		meta = UserDefinedEventMeta{
			Name:  "evnt_unkown_report",
			Level: "debug",
			Group: "sys",
		}
	}

	if level_int, err := log.ParseLevel(string(meta.Level)); err == nil {
		if level_int >= sentry_event_level {
			needReport = true
		} else {
			needReport = false
		}
	} else {
		needReport = true
	}

	return needReport, &meta
}
