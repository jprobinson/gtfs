import csv
import json
from collections import defaultdict


trains = ["1","2","3","4","5","5X","6","6X","S","L","B","D","A","G","C","E","N","Q","R","W"]

trips = dict()
for t in trains:
    trips[t] = dict()

with open('trips.txt','r') as csvin:
    reader=csv.DictReader(csvin)
    for line in reader:
        if line['route_id'] in trips:
            trips[line['trip_id']] = line['route_id']


sstop_seqs = dict()
nstop_seqs = dict()

with open('stop_times.txt','r') as csvin:
    reader=csv.DictReader(csvin)
    for line in reader:
        if line['trip_id'] in trips:
            train = trips[line['trip_id']]

            if train not in nstop_seqs:
                nstop_seqs[train] = defaultdict(list)

            if train not in sstop_seqs:
                sstop_seqs[train] = defaultdict(list)

            if line['stop_id'].endswith('S'):
                sstop_seqs[train][line['trip_id']].append([line['stop_id']])

train_stops = defaultdict(list)

for train in sstop_seqs:
    trips = sstop_seqs[train]
    for trip_id in trips:
        stops = trips[trip_id]

        if len(train_stops[train]) < len(stops):
            train_stops[train] = {'stops': stops}


all_stops = dict()

with open('stops.txt','r') as csvin:
    reader=csv.DictReader(csvin)
    # dialog = []
    for line in reader:
        # all_stops[line['stop_id']] = {"name":line['stop_name'],"lat":line['stop_lat'],"long":line['stop_lon']}
         all_stops[line['stop_id']] = line['stop_name']
        # if line['stop_id'].endswith('S') or line['stop_id'].endswith('N'):
        #    continue
        # name = line['stop_name'].replace("(",",").replace(")","")
        # names = set(name.split(","))
        # names.add(name)
        # dialog.append({"value":line['stop_id'],"synonyms":list(names)})

for train, stops in train_stops.items():
    for i, stop in enumerate(stops['stops']):
        train_stops[train]['stops'][i].append(all_stops[stop[0]])

print json.dumps(train_stops)

