.DEFAULT_GOAL := build 
.PHONY: build
build: fetch-protos protoc fetch-csvs generate

.PHONY: generate
generate: clean-generated
	@cd cmd/generate; \
	go run main.go;
	@sed -i '' -e 's/gtfs\.//g' ./nycsubwayroutes.go

.PHONY: clean-generated
clean-generated:
	@rm -rf ./nycsubwayroutes.go
	@rm -rf ./nyc-subway-routes.json
	@rm -rf ./nyc-subway-synonyms.json

.PHONY: clean
clean: clean-proto clean-csv

.PHONY: clean-proto
clean-proto:
	@rm -f gtfs-realtime.proto
	@rm -f nyct-subway.proto
	@rm -rf transit_realtime
	@mkdir transit_realtime

.PHONY: clean-csv
clean-csv:
	@rm -rf static_gtfs
	@mkdir static_gtfs

.PHONY: fetch-csvs
fetch-csvs: clean-csv
	@curl -s -o ./static_gtfs/google_transit.zip http://web.mta.info/developers/data/nyct/subway/google_transit.zip
	@cd static_gtfs; \
	unzip -q ./google_transit.zip; \
	rm -f google_transit.zip

.PHONY: fetch-protos
fetch-protos: clean-proto
	@curl -s -o gtfs-realtime.proto https://developers.google.com/transit/gtfs-realtime/gtfs-realtime.proto
	@curl -s -o nyct-subway.proto http://datamine.mta.info/sites/all/files/pdfs/nyct-subway.proto.txt
	@echo '\n\noption go_package = "transit_realtime";' >> nyct-subway.proto;

.PHONY: protoc
protoc:
	@protoc --go_out=./transit_realtime gtfs-realtime.proto
	@protoc --go_out=./transit_realtime nyct-subway.proto
