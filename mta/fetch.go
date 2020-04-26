package mta

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"

	"github.com/jprobinson/gtfs/transit_realtime"
)

type FeedType string

const (
	NumberedFeed FeedType = ""
	BlueFeed     FeedType = "-ace"
	YellowFeed   FeedType = "-nqrw"
	OrangeFeed   FeedType = "-bdfm"
	LFeed        FeedType = "-l"
	GFeed        FeedType = "-g"
	SevenFeed    FeedType = "-7"
	BrownFeed    FeedType = "-jz"
)

// GetSubwayFeed takes an API key generated from https://api.mta.info and a type
// specifying which subway feed and it will return a transit_realtime.FeedMessage with
// NYCT extensions.
func GetNYCSubwayFeed(ctx context.Context, hc *http.Client, key string, ft FeedType) (*transit_realtime.FeedMessage, error) {
	r, err := http.NewRequest(http.MethodGet,
		"https://api-endpoint.mta.info/Dataservice/mtagtfsfeeds/nyct%2Fgtfs"+string(ft), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to build request", err)
	}
	r.Header.Set("x-api-key", key)

	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to get feed", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to read feed", err)
	}

	var feed transit_realtime.FeedMessage
	err = proto.Unmarshal(body, &feed)
	if err != nil {
		return nil, fmt.Errorf("%w: unable to parse feed", err)
	}

	return &feed, nil
}
