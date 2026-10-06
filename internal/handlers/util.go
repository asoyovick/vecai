package handlers

import (
	"strconv"
	"time"
)

func timeNowDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

func formatQty(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}
