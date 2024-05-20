package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Restaurant struct {
	ID                   string
	CONTACT              string
	RESERVATION_PERIOD_A int
	RESERVATION_PERIOD_B int
	MIN_GROUP_SIZE       int
	MAX_GROUP_SIZE       int
	OPEN_HOURS_M_S       [][]string
	TEMP_STOP            []string
}

type Reservation struct {
	UID                string
	RID                string
	SID                string
	RESERVATION_STATUS string // PENDING CONFIRMED REJECTED CANCELLED
	DATE               int
	TIME               string
	Group_Size         int
}

var registered_restaurants = make(map[string]Restaurant) // ID : Restaurant Info
var reservation_logs = make(map[string]Reservation)      // RID : Reservation Info

var CURR_DAY int = 1

func getStdin() []string {
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\x00')
	lines := strings.Split(strings.TrimSpace(input), "\n")
	return lines
}

func main() {

	lines := getStdin()

	if len(lines) > 0 {

		restaurantCount, err := strconv.Atoi(lines[0])
		if err != nil {
			fmt.Println("Invalid number of restaurants:", err)
			return
		}

		lineIndex := 1
		for i := 0; i < restaurantCount; i++ {
			listResturant(lines, &lineIndex)
		}

		// Process the remaining queries
		for lineIndex < len(lines) {

			query := strings.Fields(lines[lineIndex])
			execute_query := query[0]
			if !validateQuery(execute_query) {
				fmt.Println("Invalid Query")
			}

			switch execute_query {
			case "REQUEST":
				requestReservation(query)
			case "CANCEL":
				cancelReservation(query)
			case "CONFIRM":
				confirmReservation(query)
			case "REJECT":
				rejectReservation(query)
			case "STOP":
				stopBookings(query)
			case "NEXT_DAY":
				nextDayRequest()
			case "LIST":
				lineIndex++
				listResturant(lines, &lineIndex)
				lineIndex--
			case "REMOVE":
				removeRestaurantRequest(query)
			}
			lineIndex++
		}
	}

}

// list restaurant section start

// func to list restaurants
func listResturant(lines []string, lineIndex *int) {

	var new_resturant Restaurant

	restaurant_info := strings.Fields(lines[*lineIndex])

	new_resturant.ID = restaurant_info[0]
	new_resturant.CONTACT = restaurant_info[1]
	new_resturant.RESERVATION_PERIOD_A, _ = strconv.Atoi(restaurant_info[2])
	new_resturant.RESERVATION_PERIOD_B, _ = strconv.Atoi(restaurant_info[3])
	new_resturant.MIN_GROUP_SIZE, _ = strconv.Atoi(restaurant_info[4])
	new_resturant.MAX_GROUP_SIZE, _ = strconv.Atoi(restaurant_info[5])
	*lineIndex = *lineIndex + 1

	for j := 0; j < 7; j++ {
		open_hours := strings.Split(lines[*lineIndex], " ")
		new_resturant.OPEN_HOURS_M_S = append(new_resturant.OPEN_HOURS_M_S, open_hours)
		*lineIndex = *lineIndex + 1
	}
	new_resturant.TEMP_STOP = []string{"-", "-", "-", "-", "-", "-", "-"}
	registered_restaurants[new_resturant.ID] = new_resturant

}

// list restaurant section end

// request reservation section start
func requestReservation(parts []string) {

	date, _ := strconv.Atoi(parts[4])
	grp_size, _ := strconv.Atoi(parts[6])

	var new_reservation Reservation
	new_reservation.RID = parts[1]
	new_reservation.UID = parts[2]
	new_reservation.SID = parts[3]
	new_reservation.DATE = date
	new_reservation.TIME = parts[5]
	new_reservation.Group_Size = grp_size

	if validateReservationRequest(new_reservation.RID, new_reservation.UID, new_reservation.SID, new_reservation.DATE, new_reservation.TIME, new_reservation.Group_Size) {
		new_reservation.RESERVATION_STATUS = "PENDING"
		reservation_logs[new_reservation.RID] = new_reservation
	}

}

// reservation request validator
func validateReservationRequest(
	RID string,
	UID string,
	SID string,
	Date int,
	Time string,
	Group_Size int) bool {

	if book_rest, ok := registered_restaurants[SID]; ok {
		var diff int = Date - CURR_DAY
		if diff > book_rest.RESERVATION_PERIOD_A || diff < book_rest.RESERVATION_PERIOD_B {
			fmt.Println("Error: Outside of reservation period")
			return false
		} else if Group_Size < book_rest.MIN_GROUP_SIZE || Group_Size > book_rest.MAX_GROUP_SIZE {
			fmt.Println("Error: Too many or too few people")
			return false
		} else if !checkOpeningHours(book_rest.OPEN_HOURS_M_S[(Date-1)%7], Time) {
			fmt.Println("Error: Closed")
			return false
		} else if checkOpeningHours([]string{book_rest.TEMP_STOP[(Date-1)%7]}, Time) {
			fmt.Println("Error: Reservations temporarily closed")
			return false
		} else {
			fmt.Printf("to:%s Received a reservation request: %s %s %d %s %d \n", SID, RID, UID, Date, Time, Group_Size)
		}

	} else {
		fmt.Println("Error: No such restaurant")
		return false
	}

	return true

}

func checkOpeningHours(open_hours_date []string, Time string) bool {

	h, m, err := parseTime(Time)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}
	for _, oh := range open_hours_date {
		if oh == "-" {
			continue
		}
		if isInOpeningHours(oh, h, m) {
			return true
		}
	}

	return false
}

// parseTime parses the given time string (hh:mm) and returns the hour and minute as integers.
func parseTime(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format: %s", timeStr)
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return hour, minute, nil
}

func isInOpeningHours(openingHour string, h, m int) bool {
	ranges := strings.Split(openingHour, "-")
	startTime, endTime := ranges[0], ranges[len(ranges)-1]
	sh, sm, _ := parseTime(startTime)
	eh, em, _ := parseTime(endTime)
	start := time.Date(1, 1, 1, sh, sm, 0, 0, time.UTC)
	end := time.Date(1, 1, 1, eh, em, 0, 0, time.UTC)
	given := time.Date(1, 1, 1, h, m, 0, 0, time.UTC)
	return !given.Before(start) && given.Before(end)
}

// request reservation section stop

// cancel reservation section start
func cancelReservation(parts []string) {
	var UID, RID string
	UID = parts[1]
	RID = parts[2]

	if validateCancellationRequest(UID, RID) {
		reservation_cancel := reservation_logs[RID]
		reservation_cancel.RESERVATION_STATUS = "CANCELLED"
		reservation_logs[RID] = reservation_cancel
	}

	fmt.Printf("to:%s %s has been cancelled \n", reservation_logs[RID].SID, RID)
}

func validateCancellationRequest(UID string, RID string) bool {

	if cancel_reserve, ok := reservation_logs[RID]; ok && cancel_reserve.UID == UID {
		if _, ok := registered_restaurants[cancel_reserve.SID]; !ok {
			fmt.Println("Error: No such restaurant")
			return false
		}
		if cancel_reserve.RESERVATION_STATUS == "REJECTED" {
			fmt.Println("Error: Rejected")
			return false
		}
		if cancel_reserve.RESERVATION_STATUS == "CANCELLED" {
			fmt.Println("Error: CANCELLED")
			return false
		}
		if CURR_DAY > cancel_reserve.DATE {
			fmt.Println("Error: Past reservation")
			return false
		}
		if cancel_reserve.DATE-CURR_DAY > registered_restaurants[cancel_reserve.SID].RESERVATION_PERIOD_A || cancel_reserve.DATE-CURR_DAY < registered_restaurants[cancel_reserve.SID].RESERVATION_PERIOD_A {
			fmt.Println("Please contact Restaurant", registered_restaurants[cancel_reserve.SID].CONTACT)
			return false
		}
		return true
	} else {
		fmt.Println("Error: Not found")
		return false
	}

}

// cancel reservation section stop

// confirm reservation section start
func confirmReservation(parts []string) {
	var SID, RID string
	SID = parts[1]
	RID = parts[2]

	if validateConfirmationRequest(SID, RID) {
		pending_reservation := reservation_logs[RID]
		pending_reservation.RESERVATION_STATUS = "CONFIRMED"
		reservation_logs[RID] = pending_reservation
		fmt.Printf("to:%s %s has been confirmed \n", reservation_logs[RID].UID, RID)
	}

}

func validateConfirmationRequest(SID string, RID string) bool {

	if confirm_reserve, ok := reservation_logs[RID]; ok && confirm_reserve.SID == SID {
		if _, ok := registered_restaurants[confirm_reserve.SID]; !ok {
			fmt.Println("Error: No such restaurant")
			return false
		}
		if confirm_reserve.RESERVATION_STATUS == "REJECTED" {
			fmt.Println("Error: Already rejected")
			return false
		}
		if confirm_reserve.RESERVATION_STATUS == "CONFIRMED" {
			fmt.Println("Error: Already confirmed")
			return false
		}
		if confirm_reserve.RESERVATION_STATUS == "CANCELLLED" {
			fmt.Println("Error: Already cancelled")
			return false
		}
		if CURR_DAY > confirm_reserve.DATE {
			fmt.Println("Error: Past reservation")
			return false
		}
		return true
	} else {
		fmt.Println("Error: No such reservation ID")
		return false
	}

}

// confirm reservation section stop

// reject reservation section start
func rejectReservation(parts []string) {
	var SID, RID string
	SID = parts[1]
	RID = parts[2]

	if validateRejectionRequest(SID, RID) {
		pending_reservation := reservation_logs[RID]
		pending_reservation.RESERVATION_STATUS = "REJECTED"
		reservation_logs[RID] = pending_reservation
		fmt.Printf("to:%s %s has been rejected \n", reservation_logs[RID].UID, RID)
	}
}

func validateRejectionRequest(SID string, RID string) bool {

	if reject_reserve, ok := reservation_logs[RID]; ok && reject_reserve.SID == SID {
		if _, ok := registered_restaurants[reject_reserve.SID]; !ok {
			fmt.Println("Error: No such restaurant")
			return false
		}
		if reject_reserve.RESERVATION_STATUS == "REJECTED" {
			fmt.Println("Error: Already rejected")
			return false
		}
		if reject_reserve.RESERVATION_STATUS == "CONFIRMED" {
			fmt.Println("Error: Already confirmed")
			return false
		}
		if reject_reserve.RESERVATION_STATUS == "CANCELLLED" {
			fmt.Println("Error: Already cancelled")
			return false
		}
		return true
	} else {
		fmt.Println("Error: No such reservation ID")
		return false
	}

}

// reject reservation section stop

// STOP section start
func stopBookings(parts []string) {

	var SID, Time_Range string
	var Date int

	SID = parts[1]
	Date, _ = strconv.Atoi(parts[2])
	Time_Range = parts[3]

	if validateStopRequest(SID, Date) {
		registered_restaurants[SID].TEMP_STOP[Date-1] = Time_Range
	}

}

func validateStopRequest(SID string, Date int) bool {
	if _, ok := registered_restaurants[SID]; ok {
		if Date < CURR_DAY {
			fmt.Println("Error: Specify a date today or after today")
			return false
		}
		if Date-CURR_DAY > registered_restaurants[SID].RESERVATION_PERIOD_A || Date-CURR_DAY < registered_restaurants[SID].RESERVATION_PERIOD_B {
			fmt.Println("Error: Cannot make a reservation already due to being outside the reservation period")
			return false
		}
	} else {
		fmt.Println("Error: No such restaurant")
		return false
	}
	return true
}

// STOP section end

func nextDayRequest() {
	CURR_DAY = CURR_DAY + 1

	keys := make([]string, 0, len(reservation_logs))
	for key := range reservation_logs {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	for _, RID := range keys {
		if reservation_logs[RID].RESERVATION_STATUS == "PENDING" {
			fmt.Printf("to:%s %s has been auto-rejected \n", reservation_logs[RID].UID, RID)
			pending_reserve := reservation_logs[RID]
			pending_reserve.RESERVATION_STATUS = "REJECTED"
			reservation_logs[RID] = pending_reserve
		}
	}

}

func removeRestaurantRequest(parts []string) {
	var SID string = parts[1]

	if _, ok := registered_restaurants[SID]; ok {
		delete(registered_restaurants, SID)
		// fmt.Println(SID, "removed")
	} else {
		fmt.Println("Error: No such restaurant")
	}
}

// validators
// query validator
func validateQuery(Query string) bool {

	var Queries_All = [8]string{"REQUEST", "CANCEL", "CONFIRM", "REJECT", "STOP", "NEXT_DAY", "LIST", "REMOVE"}
	for _, q := range Queries_All {
		if Query == q {
			return true
		}
	}

	return false
}
