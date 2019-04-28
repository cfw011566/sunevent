package sunevent

import (
	"math"
	"time"
)

// Reference
// https://github.com/BigZaphod/CLLocation-SunriseSunset/blob/master/CLLocation%2BSunriseSunset.m

func SunRise(latitude, longitude float64) time.Time {
	return sunRiseSet(true, latitude, longitude, 90.0)
}

func SunSet(latitude, longitude float64) time.Time {
	return sunRiseSet(false, latitude, longitude, 90.0)
}

func Dawn(latitude, longitude float64) time.Time {
	return sunRiseSet(true, latitude, longitude, 83.0)
}

func Dusk(latitude, longitude float64) time.Time {
	return sunRiseSet(false, latitude, longitude, 83.0)
}

func sunRiseSet(sunrise bool, latitude, longitude, zenith float64) time.Time {

	//zenith := 90.0
	sunset := sunrise != true
	// zenith = 83.0

	// Inputs:
	// day, month, year:      date of sunrise/sunset
	// latitude, longitude:   location for sunrise/sunset
	// zenith:                Sun's zenith for sunrise/sunset
	// offical      = 90 degrees 50'
	// civil        = 96 degrees
	// nautical     = 102 degrees
	// astronomical = 108 degrees

	// 1. first calculate the day of the year
	// N1 = floor(275 * month / 9)
	// N2 = floor((month + 9) / 12)
	// N3 = (1 + floor((year - 4 * floor(year / 4) + 2) / 3))
	// N = N1 - (N2 * N3) + day - 30

	today := time.Now()
	name, offset := today.Zone()
	loc := time.FixedZone(name, offset)
	localOffset := float64(offset) / 3600.0
	N := float64(today.YearDay())

	// 2. convert the longitude to hour value and calculate an approximate time
	// lngHour = longitude / 15

	// if rising time is desired:
	// t = N + ((6 - lngHour) / 24)
	// if setting time is desired:
	// t = N + ((18 - lngHour) / 24)

	lngHour := longitude / 15
	t := N + ((6 - lngHour) / 24)
	if sunset {
		t = N + ((18 - lngHour) / 24)
	}

	// 3. calculate the Sun's mean anomaly
	// M = (0.9856 * t) - 3.289

	M := (0.9856 * t) - 3.289

	// 4. calculate the Sun's true longitude
	// L = M + (1.916 * sin(M)) + (0.020 * sin(2 * M)) + 282.634

	L := M + (1.916 * degreeSin(M)) + (0.020 * degreeSin(2*M)) + 282.634
	// NOTE: L potentially needs to be adjusted into the range [0,360) by adding/subtracting 360
	L = normalizeRange(L, 360)

	// 5a. calculate the Sun's right ascension
	// RA = atan(0.91764 * tan(L))
	// NOTE: RA potentially needs to be adjusted into the range [0,360) by adding/subtracting 360

	RA := degreeAtan(0.91764 * degreeTan(L))
	RA = normalizeRange(RA, 360)

	// 5b. right ascension value needs to be in the same quadrant as L
	// Lquadrant  = (floor( L/90)) * 90
	// RAquadrant = (floor(RA/90)) * 90
	// RA = RA + (Lquadrant - RAquadrant)

	Lquadrant := math.Floor(L/90.0) * 90.0
	RAquadrant := math.Floor(RA/90.0) * 90.0
	RA = RA + (Lquadrant - RAquadrant)

	// 5c. right ascension value needs to be converted into hours
	// RA = RA / 15

	RA = RA / 15.0

	// 6. calculate the Sun's declination
	// sinDec = 0.39782 * sin(L)
	// cosDec = cos(asin(sinDec))

	sinDec := 0.39782 * degreeSin(L)
	cosDec := degreeCos(degreeAsin(sinDec))

	// 7a. calculate the Sun's local hour angle
	// cosH = (cos(zenith) - (sinDec * sin(latitude))) / (cosDec * cos(latitude))
	// if (cosH >  1)
	// the sun never rises on this location (on the specified date)
	// if (cosH < -1)
	// the sun never sets on this location (on the specified date)

	cosH := (degreeCos(zenith) - (sinDec * degreeSin(latitude))) / (cosDec * degreeCos(latitude))
	if cosH > 1.0 || cosH < -1.0 {
		panic("no answer")
	}

	// 7b. finish calculating H and convert into hours
	// if if rising time is desired:
	// H = 360 - acos(cosH)
	// if setting time is desired:
	// H = acos(cosH)
	// H = H / 15

	H := 360 - degreeAcos(cosH)
	if sunset {
		H = degreeAcos(cosH)
	}
	H = H / 15.0

	// 8. calculate local mean time of rising/setting
	// T = H + RA - (0.06571 * t) - 6.622

	T := H + RA - (0.06571 * t) - 6.622

	// 9. adjust back to UTC
	// UT = T - lngHour
	// NOTE: UT potentially needs to be adjusted into the range [0,24) by adding/subtracting 24

	UT := normalizeRange(T-lngHour, 24.0)

	// 10. convert UT value to local time zone of latitude/longitude
	// localT = UT + localOffset

	localT := normalizeRange(UT+localOffset, 24.0)
	hour := math.Floor(localT)
	minute := math.Floor((localT - hour) * 60.0)
	second := math.Floor(((localT-hour)*60.0 - minute) * 60.0)

	return time.Date(today.Year(), today.Month(), today.Day(), int(hour), int(minute), int(second), 0, loc)
}

func degreeToRadian(x float64) float64 {
	return (math.Pi / 180.0) * x
}

func radianToDegree(x float64) float64 {
	return (180.0 / math.Pi) * x
}

func degreeSin(x float64) float64 {
	return math.Sin(degreeToRadian(x))
}

func degreeAsin(x float64) float64 {
	return radianToDegree(math.Asin(x))
}

func degreeAtan(x float64) float64 {
	return radianToDegree(math.Atan(x))
}

func degreeTan(x float64) float64 {
	return math.Tan(degreeToRadian(x))
}

func degreeCos(x float64) float64 {
	return math.Cos(degreeToRadian(x))
}

func degreeAcos(x float64) float64 {
	return radianToDegree(math.Acos(x))
}

func normalizeRange(v, max float64) float64 {
	for v < 0 {
		v += max
	}

	for v >= max {
		v -= max
	}

	return v
}
