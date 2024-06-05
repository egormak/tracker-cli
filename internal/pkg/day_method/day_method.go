package day_method

import "time"

func IsWeekendNow() bool {
	if time.Now().Weekday().String() == "Sunday" || time.Now().Weekday().String() == "Saturday" {
		return true
	} else {
		return false
	}
}
