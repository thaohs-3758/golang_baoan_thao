package configs

import "os"

func NotificationServiceURL() string {
	if v := os.Getenv("NOTIFICATION_SERVICE_URL"); v != "" {
		return v
	}
	return "http://localhost:8081"
}
