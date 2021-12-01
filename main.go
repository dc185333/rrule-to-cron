package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/teambition/rrule-go"
)

var (
	// ***********************************************
	// DAILY
	// ***********************************************
	// rruleStr = "FREQ=DAILY;INTERVAL=1" // every day
	// rruleStr = "FREQ=DAILY;INTERVAL=5" // every 5 days

	// ***********************************************
	// WEEKLY
	// ***********************************************
	// rruleStr = "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR" // every weekday
	// rruleStr = "FREQ=WEEKLY;INTERVAL=2;BYDAY=SA,SU" // every other week on sat and sun

	// ***********************************************
	// MONTHLY
	// ***********************************************
	// rruleStr = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,2,3" // every month on the 1st, 2nd, and 3rd of the month
	rruleStr = "FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=25" // every 3 months on the 25th of the month
	// rruleStr = "FREQ=MONTHLY;INTERVAL=1;BYDAY=MO,TU;BYSETPOS=1,2" // every month on the 1st and 2nd mon and tues of the month
	// rruleStr = "FREQ=MONTHLY;INTERVAL=2;BYDAY=MO;BYSETPOS=3" // every other month on the 3rd mon of the month

	startDate = time.Date(2021, 1, 4, 23, 59, 0, 0, time.UTC)
	endDate   = time.Date(2023, 5, 30, 23, 59, 0, 0, time.UTC)
	startTime = startDate.Format("15:04")

	supportedFrequencies      = []rrule.Frequency{rrule.DAILY, rrule.WEEKLY, rrule.MONTHLY, rrule.YEARLY}
	rruleWeekdayToCronWeekday = map[rrule.Weekday]string{
		rrule.MO: "mon",
		rrule.TU: "tue",
		rrule.WE: "wed",
		rrule.TH: "thu",
		rrule.FR: "fri",
		rrule.SA: "sat",
		rrule.SU: "sun",
	}
)

func main() {
	fmt.Printf("RRULE:%s\n", rruleStr)
	fmt.Printf("startDate: %s\n", startDate)
	fmt.Printf("endDate: %s\n", endDate)

	r, err := rrule.StrToRRule(rruleStr)
	if err != nil {
		panic(err)
	}
	r.DTStart(startDate)

	if !contains(supportedFrequencies, r.Options.Freq) {
		panic(fmt.Sprintf("unsupported frequency [%s] in rrule [%s]", r.Options.Freq, r.String()))
	}

	if r.Options.Interval != 1 {
		if startDate.Before(time.Now()) {
			startDate = time.Now()
		}
		// rescheduleDate is the last day of the month before 1 year from startDate and determines when we need
		// to recalculate the date list and create new schedulers
		// eg. startDate = Nov 20 2021, endDate = Oct 31 2022
		//     startDate = Jan 1 2021, endDate = Dec 31, 2021
		rescheduleDate := time.Date(
			startDate.AddDate(1, 0, 0).Year(),
			startDate.AddDate(1, 0, 0).Month(),
			1, 0, 0, 0, 0, time.UTC).Add(-1 * time.Second)
		if endDate.After(rescheduleDate) {
			endDate = rescheduleDate
			fmt.Printf("rescheduleDate: %s\n", endDate)
		}

		dates := r.Between(
			startDate,
			endDate,
			true,
		)

		frequencies := map[string][]string{}
		var orderedMonth []string
		for _, date := range dates {
			_, m, d := date.Date()
			month := m.String()[:3]
			if _, ok := frequencies[month]; !ok {
				orderedMonth = append(orderedMonth, month)
			}
			frequencies[month] = append(frequencies[month], fmt.Sprint(d))
		}

		for _, month := range orderedMonth {
			fmt.Printf("%s of %s %s\n", strings.Join(frequencies[month], ","), month, startTime)
		}

		return
	}

	switch r.Options.Freq {
	case rrule.DAILY:
		fmt.Printf("every day %s\n", startDate.Format("15:04"))
	case rrule.WEEKLY:
		var cronDayList []string
		for _, day := range r.Options.Byweekday {
			cronDayList = append(cronDayList, rruleWeekdayToCronWeekday[day])
		}
		fmt.Printf("every %s %s\n", strings.Join(cronDayList, ","), startTime)
	case rrule.MONTHLY:
		monthDayString := ""
		var monthDayList []string
		for _, monthDay := range r.Options.Bymonthday {
			monthDayList = append(monthDayList, fmt.Sprint(monthDay))
		}
		if len(monthDayList) > 0 {
			monthDayString = strings.Join(monthDayList, ",") + " "
		}

		setPosString := ""
		var ordinalList []string
		for _, setPos := range r.Options.Bysetpos {
			ordinalList = append(ordinalList, ordinalize(setPos))
		}
		if len(ordinalList) > 0 {
			setPosString = strings.Join(ordinalList, ",") + " "
		}

		cronDayString := ""
		var cronDayList []string
		for _, day := range r.Options.Byweekday {
			cronDayList = append(cronDayList, rruleWeekdayToCronWeekday[day])
		}
		if len(cronDayList) > 0 {
			cronDayString = strings.Join(cronDayList, ",") + " "
		}

		fmt.Printf("%s%s%sof month %s\n", monthDayString, setPosString, cronDayString, startTime)
	}
}

func contains(s []rrule.Frequency, e rrule.Frequency) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ordinalize(num int) string {
	var ordinalDictionary = map[int]string{
		0: "th",
		1: "st",
		2: "nd",
		3: "rd",
		4: "th",
		5: "th",
		6: "th",
		7: "th",
		8: "th",
		9: "th",
	}

	// math.Abs() is to convert negative number to positive
	floatNum := math.Abs(float64(num))
	positiveNum := int(floatNum)

	if ((positiveNum % 100) >= 11) && ((positiveNum % 100) <= 13) {
		return strconv.Itoa(num) + "th"
	}

	return strconv.Itoa(num) + ordinalDictionary[positiveNum]
}
