package task

import (
	"fmt"
	"time"
)

func IsValidDayInMonth(year, month, day int) bool {
	if month < 1 || month > 12 || day < 1 || day > 31 {
		return false
	}
	candidate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return candidate.Day() == day
}

func NormalizeToUTCStartOfDay(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

func ApplyTimezone(cfg *RecurrenceConfig) error {
	if cfg == nil {
		return fmt.Errorf("apply timezone: config is nil")
	}

	loc := time.UTC
	if cfg.Timezone != "" {
		var err error
		loc, err = time.LoadLocation(cfg.Timezone)
		if err != nil {
			return fmt.Errorf("apply timezone: %w", err)
		}
	}

	convertToUTC := func(t time.Time) time.Time {
		y, m, d := t.Date()
		h, min, s := t.Clock()
		local := time.Date(y, m, d, h, min, s, t.Nanosecond(), loc)
		return local.UTC()
	}

	cfg.StartDate = convertToUTC(cfg.StartDate)

	if cfg.EndDate != nil {
		end := convertToUTC(*cfg.EndDate)
		cfg.EndDate = &end
	}

	if len(cfg.SpecificDates) > 0 {
		for i, d := range cfg.SpecificDates {
			cfg.SpecificDates[i] = convertToUTC(d)
		}
	}

	return nil
}

func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

func (c RecurrenceConfig) CalculateEffectiveStart(now time.Time) time.Time {
	if !c.StartDate.Before(now) {
		return c.StartDate
	}

	loc := c.resolveLocation()

	switch c.Type {
	case RecurrenceTypeDaily:
		interval := 1
		if c.Interval != nil {
			interval = *c.Interval
		}
		startDay := c.StartDate.In(loc).Truncate(24 * time.Hour)
		nowDay := now.In(loc).Truncate(24 * time.Hour)
		daysPassed := int(nowDay.Sub(startDay)/(24*time.Hour)) + 1
		steps := (daysPassed + interval - 1) / interval
		return startDay.AddDate(0, 0, steps*interval)

	case RecurrenceTypeMonthly:
		day := *c.DayOfMonth
		y, m, _ := now.In(loc).Date()
		check := time.Date(y, m, 1, c.StartDate.Hour(), c.StartDate.Minute(), 0, 0, loc)

		for {
			candidate := time.Date(check.Year(), check.Month(), day, check.Hour(), check.Minute(), 0, 0, loc)
			if candidate.Month() == check.Month() && !candidate.Before(now) {
				return candidate
			}
			check = check.AddDate(0, 1, 0)
		}

	case RecurrenceTypeSpecificDates:
		for _, d := range c.SpecificDates {
			if !d.Before(now) {
				return d.In(loc)
			}
		}
		return now.In(loc).Truncate(24 * time.Hour).AddDate(0, 0, 1)

	case RecurrenceTypeParity:
		mode := *c.ParityMode
		current := now.In(loc).Truncate(24 * time.Hour)
		if current.Before(now) {
			current = current.Add(24 * time.Hour)
		}

		for {
			isEven := current.Day()%2 == 0
			if (mode == ParityEven && isEven) || (mode == ParityOdd && !isEven) {
				return current
			}
			current = current.Add(24 * time.Hour)
		}

	default:
		return now.Truncate(24 * time.Hour)
	}
}

func (c RecurrenceConfig) resolveLocation() *time.Location {
	if c.Timezone != "" {
		if loc, err := time.LoadLocation(c.Timezone); err == nil {
			return loc
		}
	}
	if loc := c.StartDate.Location(); loc != nil {
		return loc
	}
	return time.UTC
}