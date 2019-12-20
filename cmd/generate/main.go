package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jprobinson/gtfs"
)

func main() {
	stops := writeRoutes()
	writeSynonyms(stops)
	writeStopLookup(stops)
}

func writeRoutes() map[string]gtfs.Route {
	// trip ID => route ID
	routesByTrip := getRoutesByTrip()
	// train -> trip -> stop ID
	tripsByRoute := getTripStops(routesByTrip)
	// add metadata like whoa
	// train -> route -> []stops
	stops := addStopData(tripsByRoute)

	writeGoFile("NYCSubwayRoutes", stops)
	writeJSFile("nyc-subway-routes", stops)
	return stops
}

func writeSynonyms(stops map[string]gtfs.Route) {
	var synsOut []gtfs.Synonym
	synsSeen := map[string]bool{}
	for _, stop := range getStopData() {
		if synsSeen[stop.PhoneticName] {
			continue
		}
		synsOut = append(synsOut, gtfs.Synonym{
			Value:    stop.PhoneticName,
			Synonyms: stop.Synonyms,
		})
		synsSeen[stop.PhoneticName] = true
	}
	sort.SliceStable(synsOut, func(i, j int) bool {
		return synsOut[i].Value < synsOut[j].Value
	})
	writeJSFile("nyc-subway-synonyms", synsOut)
}

func writeStopLookup(stops map[string]gtfs.Route) {
	// phono Name => Line => stop ID
	stopsOut := map[string]map[string]string{}
	for line, route := range stops {
		for _, stop := range route.Stops {
			_, exists := stopsOut[stop.PhoneticName]
			if !exists {
				stopsOut[stop.PhoneticName] = map[string]string{}
			}

			stopsOut[stop.PhoneticName][line] = stop.ID

			for _, syn := range stop.Synonyms {
				_, exists := stopsOut[syn]
				if !exists {
					stopsOut[syn] = map[string]string{}
				}
				stopsOut[syn][line] = stop.ID
			}
		}
	}
	writeGoFile("NYCSubwayStopsByName", stopsOut)
	writeJSFile("nyc-subway-stops-by-name", stopsOut)
}

func writeJSFile(name string, data interface{}) {
	jsFile, err := os.Create("../../" + strings.ToLower(name) + ".json")
	if err != nil {
		fmt.Printf("unable to open %s.json file: %s\n", name, err)
		os.Exit(1)
	}
	defer jsFile.Close()

	enc := json.NewEncoder(jsFile)
	enc.SetIndent("", "  ")
	err = enc.Encode(data)
	if err != nil {
		fmt.Println("unable to write routes.json file: ", err)
		os.Exit(1)
	}
}

func writeGoFile(name string, data interface{}) {
	goFile, err := os.Create("../../" + strings.ToLower(name) + ".go")
	if err != nil {
		fmt.Printf("unable to open %s.go file: %s\n", name, err)
		os.Exit(1)
	}
	defer goFile.Close()
	fmt.Fprintf(goFile, "package gtfs\n\nvar %s = %#v", strings.Title(name), data)
}

func addStopData(tripsByRoute map[string][]string) map[string]gtfs.Route {
	stopData := getStopData()

	out := map[string]gtfs.Route{}
	for line, route := range tripsByRoute {
		routeInfo := gtfs.Route{
			Name:       line,
			Northbound: routes[line].Northbound,
			Southbound: routes[line].Southbound,
		}

		for _, stopID := range route {
			stop, ok := stopData[stopID]
			if !ok {
				fmt.Printf("we didnt find data for %s!", stopID)
				os.Exit(1)
			}

			routeInfo.Stops = append(routeInfo.Stops, stop)
		}

		out[line] = routeInfo
	}

	// add transfer datas

	for line, route := range out {
		for i, stop := range route.Stops {
			var trans []gtfs.Transfer
			// find any other lines that also have this stop via same ID
			// N^2 yall!!! :shruggie:
			for _, route2 := range out {
				if route2.Name == route.Name {
					continue
				}
				for _, stop2 := range route2.Stops {
					if stop2.ID != stop.ID {
						continue
					}
					trans = append(trans, gtfs.Transfer{
						StopID: stop.ID,
						Route:  route2.Name,
					})
				}
			}
			// find any other lines that DO NOT have this stop via same ID
			// we should already have these ref'd just need to add datas
			for _, xfer := range stop.Transfers {
				if xfer.StopID == stop.ID {
					continue
				}
				for _, route2 := range out {
					if route2.Name == route.Name {
						continue
					}
					for _, stop2 := range route2.Stops {
						if stop2.ID != xfer.StopID {
							continue
						}
						trans = append(trans, gtfs.Transfer{
							StopID: xfer.StopID,
							Route:  route2.Name,
						})
					}
				}
			}

			sort.Slice(trans, func(i, j int) bool {
				return trans[i].Route < trans[j].Route
			})
			out[line].Stops[i].Transfers = trans
		}
	}

	return out
}

func getStopData() map[string]gtfs.Stop {
	stopsFile, err := os.Open("../../static_gtfs/stops.txt")
	if err != nil {
		fmt.Println("unable to open stops.txt:", err)
		os.Exit(1)
	}
	defer stopsFile.Close()

	cols := map[string]int{}
	stopData := map[string]gtfs.Stop{}
	r := csv.NewReader(stopsFile)
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("unable to read stops.txt:", err)
			os.Exit(1)
		}
		if len(cols) == 0 {
			for idx, val := range record {
				cols[val] = idx
			}
			continue
		}

		stopID := record[cols["stop_id"]]
		mtaName := record[cols["stop_name"]]
		displayName, phonoName, syns := makeStopNames(mtaName)
		stopData[stopID] = gtfs.Stop{
			ID:           stopID,
			MTAName:      mtaName,
			DisplayName:  displayName,
			PhoneticName: phonoName,
			Synonyms:     syns,
		}
	}

	// now add transfers!
	transFile, err := os.Open("../../static_gtfs/transfers.txt")
	if err != nil {
		fmt.Println("unable to open transfers.txt:", err)
		os.Exit(1)
	}
	defer transFile.Close()

	cols = map[string]int{}
	tr := csv.NewReader(transFile)
transloop:
	for {
		record, err := tr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("unable to read transfers.txt:", err)
			os.Exit(1)
		}
		if len(cols) == 0 {
			for idx, val := range record {
				cols[val] = idx
			}
			continue
		}

		from := record[cols["from_stop_id"]]
		to := record[cols["to_stop_id"]]
		if to == from {
			continue transloop
		}

		for _, xfer := range stopData[from].Transfers {
			if xfer.StopID == to {
				continue transloop
			}
		}
		fromStop := stopData[from]
		fromStop.Transfers = append(fromStop.Transfers, gtfs.Transfer{StopID: to})
		stopData[from] = fromStop
	}
	return stopData
}

func getTripStops(trips map[string]string) map[string][]string {
	stopsFile, err := os.Open("../../static_gtfs/stop_times.txt")
	if err != nil {
		fmt.Println("unable to open stop_times.txt:", err)
		os.Exit(1)
	}
	defer stopsFile.Close()

	cols := map[string]int{}
	// train -> trip -> stops
	northTrips := map[string]map[string][]string{}
	southTrips := map[string]map[string][]string{}
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
		if _, ok := northTrips[train]; !ok {
			northTrips[train] = map[string][]string{}
		}
		if _, ok := southTrips[train]; !ok {
			southTrips[train] = map[string][]string{}
		}

		if strings.HasSuffix(record[cols["stop_id"]], "S") {
			stopID := strings.TrimSuffix(record[cols["stop_id"]], "S")
			southTrips[train][record[cols["trip_id"]]] = append(southTrips[train][record[cols["trip_id"]]],
				stopID)
		} else {
			stopID := strings.TrimSuffix(record[cols["stop_id"]], "N")
			northTrips[train][record[cols["trip_id"]]] = append(northTrips[train][record[cols["trip_id"]]],
				stopID)
		}
	}

	// for each train, find the longest trip route
	trainTrips := map[string][]string{}
	for route, _ := range routes {
		max := len(trainTrips[route])

		for _, trips := range northTrips[route] {
			if len(trips) > max {
				for i := len(trips)/2 - 1; i >= 0; i-- {
					opp := len(trips) - 1 - i
					trips[i], trips[opp] = trips[opp], trips[i]
				}
				trainTrips[route] = trips
				max = len(trips)
			}
		}
		for _, trips := range southTrips[route] {
			if len(trips) > max {
				trainTrips[route] = trips
			}
		}
	}
	return trainTrips
}

func getRoutesByTrip() map[string]string {
	tripsFile, err := os.Open("../../static_gtfs/trips.txt")
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

var (
	reAvenue  = regexp.MustCompile("(Av)($| - |,| )")
	reAvenues = regexp.MustCompile("(Avs)($| - |,| )")
	reStreet  = regexp.MustCompile("(St)($| - |,| )")
	reStreets = regexp.MustCompile("(Sts)($| - |,| )")
	rePlace   = regexp.MustCompile("(Pl)($| - |,| )")
)

func makeStopNames(mtaName string) (string, string, []string) {
	replacements := [][]string{
		{" (", ", "},
		{")", ""},
		{"Hts", "Heights"},
		{"Sq", "Square"},
		{"Pkwy", "Parkway"},
		{"Blvd", "Boulevard"},
		{"Hwy", "Highway"},
		{"Ctr", "Center"},
		{"Jct", "Junction"},
		{"1 ", "1st "},
		{"2 ", "2nd "},
		{"2-", "2nd "},
		{"3 ", "3rd "},
		{"4 ", "4th "},
		{"4-", "4th "},
		{"5 ", "5th "},
		{"6 ", "6th "},
		{"7 ", "7th "},
		{"7-", "7th "},
		{"8 ", "8th "},
		{"9 ", "9th "},
		{"0 ", "0th "},
		{"E ", "East "},
		{"W ", "West "},
		{"N ", "North "},
		{"S ", "South "},
		{"/", ", "},
	}

	displayName := mtaName
	for _, rep := range replacements {
		displayName = strings.ReplaceAll(displayName, rep[0], rep[1])
	}

	displayName = fixStAve(reAvenue, displayName, "Avenue")
	displayName = fixStAve(reAvenues, displayName, "Avenues")
	displayName = fixStAve(reStreet, displayName, "Street")
	displayName = fixStAve(reStreets, displayName, "Streets")
	displayName = fixStAve(rePlace, displayName, "Place")
	displayName = strings.Join(strings.Fields(displayName), " ")

	syn := strings.ReplaceAll(displayName, " - ", ", ")
	syns := strings.Split(syn, ",")
	for i, syn := range syns {
		syns[i] = strings.TrimSpace(syn)
	}
	if len(syns) == 2 {
		syns = append(syns, syns[1]+", "+syns[0])
	}
	if syn != displayName {
		syns = append(syns, syn)
	}
	return displayName, syn, syns
}

func fixStAve(re *regexp.Regexp, given, want string) string {
	return re.ReplaceAllStringFunc(given, func(found string) string {
		rep := want
		if strings.Contains(found, "-") || strings.Contains(found, ",") {
			rep += ","
		}
		rep += " "
		return rep
	})
}

var routes = map[string]gtfs.Route{
	"1":  {Northbound: "Bronx", Southbound: "South Ferry"},
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
