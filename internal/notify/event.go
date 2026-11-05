package notify

import "time"

// Level matches events.level and channel min_level.
type Level string

const (
	LevelError   Level = "error"
	LevelWarning Level = "warning"
	LevelSuccess Level = "success"
	LevelInfo    Level = "info"
)

// State is used for stateful alert types.
type State string

const (
	StateFiring   State = "firing"
	StateResolved State = "resolved"
)

// Well-known event types (typed events only are delivered externally).
const (
	TypeDeployFailed              = "deploy.failed"
	TypeDeploySucceeded           = "deploy.succeeded"
	TypeBackupFailed              = "backup.failed"
	TypeBackupOverdue             = "backup.overdue"
	TypeBackupSucceeded           = "backup.succeeded"
	TypeDiskHigh                  = "disk.high"
	TypeCertExpiring              = "cert.expiring"
	TypeNotificationChannelFailed = "notification.channel_failed"
	TypeTestPing                  = "test.ping"
)

// Event is a typed platform alert for external channels.
type Event struct {
	Type         string
	Level        Level
	State        State
	DedupKey     string
	Title        string
	Detail       string
	ProjectID    string
	ServiceID    string
	DeploymentID string
	Href         string
	Data         map[string]any
}

// DefaultExternalEventTypes are pre-selected when creating a channel (deploy.succeeded off).
func DefaultExternalEventTypes() []string {
	return []string{
		TypeDeployFailed,
		TypeBackupFailed,
		TypeBackupOverdue,
		TypeDiskHigh,
		TypeCertExpiring,
	}
}

func categoryForType(t string) string {
	switch {
	case t == TypeDeployFailed || t == TypeDeploySucceeded:
		return "deploys"
	case t == TypeBackupFailed || t == TypeBackupOverdue || t == TypeBackupSucceeded:
		return "backups"
	case t == TypeCertExpiring:
		return "certificates"
	case t == TypeDiskHigh:
		return "system"
	default:
		return "system"
	}
}

func isStatefulType(t string) bool {
	switch t {
	case TypeBackupFailed, TypeBackupOverdue, TypeBackupSucceeded, TypeDiskHigh, TypeCertExpiring, TypeNotificationChannelFailed:
		return true
	default:
		return false
	}
}

func levelRank(l Level) int {
	switch l {
	case LevelInfo:
		return 1
	case LevelSuccess:
		return 2
	case LevelWarning:
		return 3
	case LevelError:
		return 4
	default:
		return 0
	}
}

// LevelMeetsMin returns true when ev meets or exceeds channel min_level.
func LevelMeetsMin(min Level, ev Level) bool {
	return levelRank(ev) >= levelRank(min)
}

const defaultReminderInterval = 24 * time.Hour
