package main

import "time"

type timeWithZone struct {
	tzone *time.Location
}

func newTimeWithZone() *timeWithZone {
	t := timeWithZone{
		tzone: time.FixedZone("Asia/Tokyo", 9*60*60),
	}

	return &t
}

func (t *timeWithZone) beginningOfDay() time.Time {
	now := time.Now().UTC().In(t.tzone)

	year := now.Year()
	month := now.Month()
	day := now.Day()

	return time.Date(year, month, day-1, 0, 0, 0, 0, t.tzone)
}

func (t *timeWithZone) endOfDay() time.Time {
	now := time.Now().UTC().In(t.tzone)

	year := now.Year()
	month := now.Month()
	day := now.Day()

	return time.Date(year, month, day-1, 23, 59, 59, 59, t.tzone)
}
