package utility

import (
	"time"
)

// Age computes the age from a user's DOB.
func Age(dob time.Time) int {
	return AgeAt(dob, time.Now())
}

// AgeAt computes the user's age at a given date/time.
func AgeAt(dob, now time.Time) int {
	// How old they will turn by the end of this year.
	var age = now.Year() - dob.Year()

	// If their month hasn't arrived, subtract one.
	if now.Month() < dob.Month() {
		age--
	} else if now.Month() == dob.Month() {
		// In their birth month, has their day come?
		if dob.Day() < now.Day() {
			age--
		}
	}

	return age
}
