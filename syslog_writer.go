package logger

import (
	"fmt"
	"net"
	"os"
	"time"
)

type SyslogFacility int

const (
	LOG_F_KERN SyslogFacility = iota
	LOG_F_USER
	LOG_F_MAIL
	LOG_F_DAEMON
	LOG_F_AUTH
	LOG_F_SYSLOG
	LOG_F_LPR
	LOG_F_NEWS
	LOG_F_UUCP
	LOG_F_CRON
	LOG_F_AUTHPRIV
	LOG_F_FTP
	LOG_F_NTP
	LOG_F_AUDIT
	LOG_F_F_ALERT
	LOG_F_CLOCK
	LOG_F_LOCAL0
	LOG_F_LOCAL1
	LOG_F_LOCAL2
	LOG_F_LOCAL3
	LOG_F_LOCAL4
	LOG_F_LOCAL5
	LOG_F_LOCAL6
	LOG_F_LOCAL7
)

type SyslogSeverity int

const (
	LOG_S_EMERG SyslogSeverity = iota
	LOG_S_ALERT
	LOG_S_CRIT
	LOG_S_ERR
	LOG_S_WARNING
	LOG_S_NOTICE
	LOG_S_INFO
	LOG_S_DEBUG
)

type SyslogConfig struct {
	Facility SyslogFacility
	Severity SyslogSeverity
	Hostname string
	Addr     string
	Protocol string
	Tag      string
}

type SyslogWriter struct {
	config *SyslogConfig
	conn   net.Conn
}

func (w *SyslogWriter) Write(p []byte) (int, error) {
	pri := int(w.config.Facility*8) + int(w.config.Severity)
	timestamp := time.Now().Format(time.RFC3339)
	msg := fmt.Sprintf("<%d>1 %s %s %s - - - %s\n", pri, timestamp, w.config.Hostname, w.config.Tag, p)
	fmt.Println("syslog msg:", msg)
	return w.conn.Write([]byte(msg))
}

func NewConfigSyslogWriter(config *SyslogConfig) (*SyslogWriter, error) {
	conn, err := net.Dial(config.Protocol, config.Addr)
	if err != nil {
		return nil, err
	}
	return &SyslogWriter{
		config: config,
		conn:   conn,
	}, nil
}

func NewSyslogWriter(addr string) (*SyslogWriter, error) {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "unknown"
	}
	config := &SyslogConfig{
		Facility: LOG_F_LOCAL1,
		Severity: LOG_S_INFO,
		Hostname: hostname,
		Addr:     addr,
		Protocol: "udp",
		Tag:      hostname,
	}
	return NewConfigSyslogWriter(config)
}
