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
	rruleStr = "FREQ=DAILY;INTERVAL=1" // every day
	// rruleStr = "FREQ=DAILY;INTERVAL=5" // every 5 days

	// ***********************************************
	// WEEKLY
	// ***********************************************
	// Note: Weekday is denoted as MO,TU,WE,TH,FR
	//       Weekend is denoted as SA,SU
	// rruleStr = "FREQ=WEEKLY;INTERVAL=1;BYDAY=MO,TU,WE,TH,FR" // every weekday
	// rruleStr = "FREQ=WEEKLY;INTERVAL=4;BYDAY=SA,SU" // every 4 weeks on sat and sun (weekend)

	// ***********************************************
	// MONTHLY
	// ***********************************************
	// rruleStr = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=1,2,3" // every month on the 1st, 2nd, and 3rd of the month
	// rruleStr = "FREQ=MONTHLY;INTERVAL=3;BYMONTHDAY=25" // every 3 months on the 25th of the month
	// rruleStr = "FREQ=MONTHLY;INTERVAL=1;BYDAY=MO,TU;BYSETPOS=1,2" // every month on the 1st and 2nd mon and tues of the month
	// rruleStr = "FREQ=MONTHLY;INTERVAL=2;BYDAY=MO;BYSETPOS=-1" // every other month on the last mon of the month

	// ***********************************************
	// YEARLY
	// ***********************************************
	// rruleStr = "FREQ=YEARLY;INTERVAL=1;BYMONTH=1,2;BYMONTHDAY=1,2,3" // every year on the 1st, 2nd, and 3rd of Jan and Feb
	// rruleStr = "FREQ=YEARLY;INTERVAL=2;BYMONTH=1;BYMONTHDAY=1" // every other year on Jan 1st
	// rruleStr = "FREQ=YEARLY;INTERVAL=1;BYMONTH=1,2;BYDAY=MO;BYSETPOS=1,2" // every year on the 1st and 2nd Mon in Jan and Feb
	// rruleStr = "FREQ=YEARLY;INTERVAL=2;BYMONTH=3;BYDAY=MO;BYSETPOS=3" // every other year on the 3rd Mon of Mar

	startDate = time.Date(2021, 11, 8, 23, 59, 0, 0, time.UTC)
	endDate   = time.Date(2022, 12, 30, 23, 59, 0, 0, time.UTC)
	startTime = startDate.Format("15:04")

	// rescheduleFreq defines how often we should reschedule interval calculations, it is used to calculate the rescheduleDate
	// which is one second before the 1st of the month of startDate + rescheduleFreq
	rescheduleFreq = rescheduleFrequency{
		years:  1,
		months: 0,
		days:   0,
	}

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

type rescheduleFrequency struct {
	years, months, days int
}

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

	// INTERVAL>1 or BYSETPOS contains -1 (indicating "last" ordinal)
	if r.Options.Interval > 1 || containsNegative(r.Options.Bysetpos) {
		if startDate.Before(time.Now()) {
			startDate = time.Now()
		}
		// rescheduleDate is the last day of the month before startDate + reschedFreq and determines when we need
		// to recalculate the date list and create new schedulers
		// eg. startDate = Nov 20 2021, reschedulFre = 1 year, rescheduleDate = Oct 31 2022
		//     startDate = Jan 1 2021, rescheduleFreq = 1 month, rescheduleDate = Jan 31, 2021
		rescheduleDate := time.Date(
			startDate.AddDate(rescheduleFreq.years, rescheduleFreq.months, rescheduleFreq.days).Year(),
			startDate.AddDate(rescheduleFreq.years, rescheduleFreq.months, rescheduleFreq.days).Month(),
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
		var orderedMonthKeys []string
		for _, date := range dates {
			_, m, d := date.Date()
			month := m.String()[:3]
			if _, ok := frequencies[month]; !ok {
				orderedMonthKeys = append(orderedMonthKeys, month)
			}
			frequencies[month] = append(frequencies[month], fmt.Sprint(d))
		}

		// uniqueDayList combines months with the same dayList
		uniqueDayList := map[string][]string{}
		var orderedUniqueDayListKeys []string
		for _, month := range orderedMonthKeys {
			dayList := strings.Join(frequencies[month], ",")
			if _, ok := uniqueDayList[dayList]; !ok {
				orderedUniqueDayListKeys = append(orderedUniqueDayListKeys, dayList)
			}
			uniqueDayList[dayList] = append(uniqueDayList[dayList], month)
		}

		for _, dayList := range orderedUniqueDayListKeys {
			fmt.Printf("%s of %s %s\n", dayList, strings.Join(uniqueDayList[dayList], ","), startTime)
		}

		return
	}

	// INTERVAL=1 and BYSETPOS does not contain -1
	switch r.Options.Freq {
	case rrule.DAILY:
		fmt.Printf("every day %s\n", startDate.Format("15:04"))
	case rrule.WEEKLY:
		fmt.Println(constructWeeklyString(r))
	case rrule.MONTHLY, rrule.YEARLY:
		fmt.Println(constructMonthlyAndYearlyString(r))
	default:
		panic(fmt.Sprintf("unsupported frequency [%s] in rrule [%s]", r.Options.Freq, r.String()))
	}
}

func constructWeeklyString(r *rrule.RRule) string {
	var cronDayList []string
	for _, day := range r.Options.Byweekday {
		cronDayList = append(cronDayList, rruleWeekdayToCronWeekday[day])
	}

	return fmt.Sprintf("every %s %s\n", strings.Join(cronDayList, ","), startTime)
}

func constructMonthlyAndYearlyString(r *rrule.RRule) string {
	monthString := "month"
	var monthList []string
	for _, month := range r.Options.Bymonth {
		monthList = append(monthList, time.Month(month).String()[:3])
	}
	if len(monthList) > 0 {
		monthString = strings.Join(monthList, ",")
	}

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

	return fmt.Sprintf("%s%s%sof %s %s", monthDayString, setPosString, cronDayString, monthString, startTime)
}

func contains(s []rrule.Frequency, e rrule.Frequency) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func containsNegative(s []int) bool {
	for _, v := range s {
		if v == -1 {
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
