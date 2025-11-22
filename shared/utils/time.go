package utils

import "time"

func NowUTC() time.Time {
    return time.Now().UTC()
}

func FormatISO8601(t time.Time) string {
    return t.UTC().Format(time.RFC3339)
}

func ParseISO8601(value string) (time.Time, error) {
    return time.Parse(time.RFC3339, value)
}