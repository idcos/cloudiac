package utils

import "time"

// LastDaysMidnight 基于 now，获取近 n 天前的 0 点(含当天)
// 假设 now 为 2021-9-17 18:18:00:
//   - 若 n 为 1 则返回, 2021-09-17 00:00:00
//   - 若 n 为 2 则返回, 2021-09-16 00:00:00
func LastDaysMidnight(n int, nows ...time.Time) time.Time {
	if n <= 0 {
		panic("days must be greater than 0")
	}

	var now time.Time
	if len(nows) > 0 {
		now = nows[0]
	} else {
		now = time.Now()
	}

	y, m, d := now.AddDate(0, 0, -(n - 1)).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}
