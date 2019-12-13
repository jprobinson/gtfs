package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {

	// trip ID => route ID
	// trips.txt
	routesByTrip := getRoutesByTrip()

	// train -> trip -> stop ID
	// stop_times.txt
	north, _ := getTripStops(routesByTrip)
	var max []string
	for _, stops := range north["L"] {
		if len(stops) > len(max) {
			fmt.Println("NEW MAX!")
			max = stops
			fmt.Println(max)
		}
	}
	//	fmt.Printf("\n\n%#v", south["L"])
}

func getTripStops(trips map[string]string) (north map[string][]string, south map[string][]string) {
	stopsFile, err := os.Open("./static_gtfs/stop_times.txt")
	if err != nil {
		fmt.Println("unable to open stop_times.txt:", err)
		os.Exit(1)
	}
	defer stopsFile.Close()

	cols := map[string]int{}
	// train -> trip -> stops
	north = map[string]map[string][]string{}
	south = map[string]map[string][]string{}
	r := csv.NewReader(stopsFile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("unable to read stop_times.txt:", err)
			os.Exit(1)
		}
		if len(cols) == 0 {
			for idx, val := range record {
				cols[val] = idx
			}
			continue
		}

		train := trips[record[cols["trip_id"]]]
		if _, ok := north[train]; !ok {
			north[train] = map[string][]string{}
		}
		if _, ok := south[train]; !ok {
			south[train] = map[string][]string{}
		}

		if strings.HasSuffix(record[cols["stop_id"]], "S") {
			south[train][record[cols["trip_id"]]] = append(south[train][record[cols["trip_id"]]],
				record[cols["stop_id"]])
		} else {
			north[train][record[cols["trip_id"]]] = append(north[train][record[cols["trip_id"]]],
				record[cols["stop_id"]])
		}
	}

	return north, south
}

func getRoutesByTrip() map[string]string {
	tripsFile, err := os.Open("./static_gtfs/trips.txt")
	if err != nil {
		fmt.Println("unable to open trips.txt:", err)
		os.Exit(1)
	}
	defer tripsFile.Close()

	cols := map[string]int{}
	routesByTrip := map[string]string{}
	r := csv.NewReader(tripsFile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("unable to read trips.txt:", err)
			os.Exit(1)
		}
		if len(cols) == 0 {
			for idx, val := range record {
				cols[val] = idx
			}
			continue
		}

		route := record[cols["route_id"]]
		if _, ok := routes[route]; ok {
			routesByTrip[record[cols["trip_id"]]] = route
		}
	}
	return routesByTrip
}

type (
	Route struct {
		Name string

		Northbound string
		Southbound string

		Trips [][]Stop
	}
	Stop struct {
		ID       string
		Name     string
		Synonyms []string
	}
)

var routes = map[string]Route{
	"1":  {Northbound: "Bronx", Southbound: "South&nbsp;Ferry"},
	"2":  {Northbound: "Bronx", Southbound: "Brooklyn"},
	"3":  {Northbound: "Harlem", Southbound: "Brooklyn"},
	"4":  {Northbound: "Bronx", Southbound: "Brooklyn"},
	"5":  {Northbound: "Bronx", Southbound: "Brooklyn"},
	"5X": {Northbound: "Bronx", Southbound: "Brooklyn"},
	"6":  {Northbound: "Bronx", Southbound: "Brooklyn Brdg"},
	"7":  {Northbound: "Queens", Southbound: "Manhattan"},
	"6X": {Northbound: "Bronx", Southbound: "Brooklyn Brdg"},
	"S":  {Northbound: "", Southbound: ""},
	"L":  {Northbound: "Manhattan", Southbound: "Brooklyn"},
	"B":  {Northbound: "Bronx", Southbound: "Brooklyn"},
	"D":  {Northbound: "Bronx", Southbound: "Brooklyn"},
	"A":  {Northbound: "Manhattan", Southbound: "Queens"},
	"G":  {Northbound: "Queens", Southbound: "Brooklyn"},
	"C":  {Northbound: "Manhattan", Southbound: "Brooklyn"},
	"E":  {Northbound: "Queens", Southbound: "Manhattan"},
	"N":  {Northbound: "Manhattan", Southbound: "Brooklyn"},
	"Q":  {Northbound: "Manhattan", Southbound: "Brooklyn"},
	"R":  {Northbound: "Queens", Southbound: "Brooklyn"},
	"W":  {Northbound: "Queens", Southbound: "Manhattan"},
	"J":  {Northbound: "Queens", Southbound: "Manhattan"},
	"F":  {Northbound: "Queens", Southbound: "Brooklyn"},
	"M":  {Northbound: "Queens", Southbound: "Brooklyn"},
	"Z":  {Northbound: "Queens", Southbound: "Manhattan"},
}
