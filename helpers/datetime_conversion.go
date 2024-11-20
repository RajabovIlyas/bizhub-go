package helpers

import "time"

// time.Now() => returns UTC time, not Ashgabat time.
// so to keep every time instance in Ashgabat time, use this function.
// front end should send time in either d.toJSON() format, or
// d.toISOString() format, both are the same.
func DTForAshgabat(dt time.Time) (time.Time, error) { // Ashgabat wagtyna owuryar
	location, err := time.LoadLocation("Asia/Ashgabat")
	if err != nil {
		return dt, err
	}
	return dt.In(location), nil
}

func UTCDateFromISOString(isoString string) (time.Time, error) { // dine senani almak ucin
	t, err := time.Parse("02-01-2006T15:04:05.999Z", isoString)
	if err != nil {
		return TodayAsDate(), err
	}
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC), nil
}

func StringToDate(isoString string) (time.Time, error) { // dine senani almak ucin
	t, err := time.Parse("02-01-2006", isoString)
	if err != nil {
		return TodayAsDate(), err
	}
	return t, nil
}
func TodayAsDate() time.Time { // dine senani almak ucin
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return today
}
