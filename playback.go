package malgova

import (
	"time"

	"github.com/sivamgr/kstreamdb"
)

//PlaybackFeed struct
type PlaybackFeed struct {
	dateToPlay time.Time
	db         kstreamdb.DB
}

// Run PlaybackFeed
func (f *PlaybackFeed) Run(fCallback func(t kstreamdb.TickData)) {
	f.db.PlaybackDate(f.dateToPlay, fCallback)
}

// Setup PlaybackFeed
func (f *PlaybackFeed) Setup(kdbPath string) {
	f.db = kstreamdb.SetupDatabase("/home/pi/data-kbridge/data/")
	f.dateToPlay = time.Now()
}

// SetDate for PlaybackFeed
func (f *PlaybackFeed) SetDate(dt time.Time) {
	f.dateToPlay = dt
}
