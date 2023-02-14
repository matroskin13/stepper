package stepper

import (
	"time"

	"github.com/teambition/rrule-go"
)

const (
	Monday    = 0
	Tuesday   = 1
	Wednesday = 2
	Thursday  = 3
	Friday    = 4
	Saturday  = 5
	Sunday    = 6
)

type Schedule struct {
	freq     rrule.Frequency
	interval int
	days     []rrule.Weekday
	hours    []int
	minutes  []int
}

func (s *Schedule) toRRule() (*rrule.RRule, error) {
	r, err := rrule.NewRRule(rrule.ROption{
		Freq:      s.freq,
		Interval:  s.interval,
		Byweekday: s.days,
		Byhour:    s.hours,
		Dtstart:   time.Now(),
		Byminute:  s.minutes,
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *Schedule) Interval(interval int) *Schedule {
	s.interval = interval

	return s
}

func (s *Schedule) AtDays(days ...int) *Schedule {
	for _, day := range days {
		s.days = append(s.days, map[int]rrule.Weekday{
			0: rrule.MO,
			1: rrule.TH,
			2: rrule.WE,
			3: rrule.TH,
			4: rrule.FR,
			5: rrule.SA,
			6: rrule.SU,
		}[day])
	}

	return s
}

func (s *Schedule) AtHours(hours ...int) *Schedule {
	s.hours = hours
	return s
}

func (s *Schedule) AtMinutes(minutes ...int) *Schedule {
	s.minutes = minutes
	return s
}

func EveryDay() *Schedule {
	return &Schedule{
		freq:     rrule.DAILY,
		interval: 1,
	}
}

func EveryMonth() *Schedule {
	return &Schedule{
		freq:     rrule.MONTHLY,
		interval: 1,
	}
}

func EveryWeek() *Schedule {
	return &Schedule{
		freq:     rrule.WEEKLY,
		interval: 1,
	}
}

func EveryHour() *Schedule {
	return &Schedule{
		freq:     rrule.HOURLY,
		interval: 1,
	}
}

func EverySecond() *Schedule {
	return &Schedule{
		freq:     rrule.SECONDLY,
		interval: 1,
	}
}
