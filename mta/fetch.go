package mta

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/golang/protobuf/proto"

	"github.com/jprobinson/gtfs/transit_realtime"
)

type FeedType int

const (
	NumberedFeed FeedType = 1
	BlueFeed     FeedType = 26
	YellowFeed   FeedType = 16
	OrangeFeed   FeedType = 21
	LFeed        FeedType = 2
	GFeed        FeedType = 31
	SevenFeed    FeedType = 51
	BrownFeed    FeedType = 36
)

// GetSubwayFeed takes an API key generated from http://datamine.mta.info/user/register
// and a boolean specifying which feed (1,2,3,4,5,6,S trains OR L train) and
// it will return a transit_realtime.FeedMessage with NYCT extensions.
func GetNYCSubwayFeed(ctx context.Context, hc *http.Client, key string, ft FeedType) (*transit_realtime.FeedMessage, error) {
	url := "http://datamine.mta.info/mta_esi.php?key=" + key +
		"&feed_id=" + strconv.Itoa(int(ft))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("unable to get feed: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read feed: %w", err)
	}

	var feed transit_realtime.FeedMessage
	err = proto.Unmarshal(body, &feed)
	if err != nil {
		return nil, fmt.Errorf("unable to parse feed: %w", err)
	}

	return &feed, nil
}
