package mta

import (
	"sort"
	"strings"
	"time"

	"github.com/jprobinson/gtfs/transit_realtime"
)

// Trains will accept a stopId plus a train line (found here: http://web.mta.info/developers/data/nyct/subway/google_transit.zip)
// and returns a list of updates from northbound and southbound trains
func Trains(f *transit_realtime.FeedMessage, stopId, line string) (alerts []*transit_realtime.Alert, northbound, southbound []*transit_realtime.TripUpdate_StopTimeUpdate) {
	for _, ent := range f.Entity {
		if ent.TripUpdate != nil {
			for _, upd := range ent.TripUpdate.StopTimeUpdate {
				if strings.HasPrefix(*upd.StopId, stopId) &&
					line == *ent.TripUpdate.Trip.RouteId {
					if strings.HasSuffix(*upd.StopId, "N") {
						northbound = append(northbound, upd)
					} else if strings.HasSuffix(*upd.StopId, "S") {
						southbound = append(southbound, upd)
					}
				}
			}
			if ent.Alert != nil {
				alerts = append(alerts, ent.Alert)
			}
		}

	}
	return alerts, northbound, southbound
}

// FeedNextTrainTimes will return an ordered slice of upcoming train departure times
// in either direction for a specific feed.
func FeedNextTrainTimes(f *transit_realtime.FeedMessage, stopId, line string) (alerts []*transit_realtime.Alert, northbound, southbound []time.Time) {
	alerts, north, south := Trains(f, stopId, line)
	northbound = NextTrainTimes(north)
	southbound = NextTrainTimes(south)
	return alerts, northbound, southbound
}

// NextTrainTimes will extract the departure times from the given
// update slice, order and return them.
func NextTrainTimes(updates []*transit_realtime.TripUpdate_StopTimeUpdate) []time.Time {
	var times []time.Time

	for _, upd := range updates {
		if upd.Departure == nil {
			continue
		}
		unix := *upd.Departure.Time
		if upd.Departure.Delay != nil {
			unix += int64(*upd.Departure.Delay * 1000)
		}
		dept := time.Unix(unix, 0)
		if dept.After(time.Now()) {
			times = append(times, dept)
		}
	}
	sort.Sort(timeSlice(times))
	if len(times) > 5 {
		times = times[:5]
	}
	return times
}

type timeSlice []time.Time

func (t timeSlice) Len() int {
	return len(t)
}

func (t timeSlice) Less(i, j int) bool {
	return t[i].Before(t[j])
}

func (t timeSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
