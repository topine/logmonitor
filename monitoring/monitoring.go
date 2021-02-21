package monitoring

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/orcaman/concurrent-map"
	"github.com/papertrail/go-tail/follower"
)

type HttpLogMonitoring struct {
	CollectionInterval int
	CollectionChannel  chan Metrics
	AlertInterval      int
	AlertThreshold     float64
	AlertChannel       chan string
	LogFilename        string
	returnCodeMap      cmap.ConcurrentMap
	hitMap             cmap.ConcurrentMap
	bytesMap           cmap.ConcurrentMap
	oldBytesMap        cmap.ConcurrentMap
	oldHitMap          cmap.ConcurrentMap
	previousTotalHits  uint64
	alertTriggered     bool
	follower           *follower.Follower
}

type Section struct {
	Name string
	Hits uint64
}

type Metrics struct {
	TotalHits       uint64
	TotalBytes      uint64
	TopSectionsHits []Section
	ReturnCodes     map[uint]uint64
}

func (m *HttpLogMonitoring) Monitor() {

	m.hitMap = cmap.New()
	m.oldHitMap = cmap.New()
	m.returnCodeMap = cmap.New()
	m.bytesMap = cmap.New()
	m.oldBytesMap = cmap.New()

	go func() {
		gocron.Every(10).Seconds().Do(m.scrap)
		gocron.Every(120).Seconds().Do(m.alert, m.AlertThreshold)
		<-gocron.Start()
	}()

	var err error
	m.follower, err = follower.New(m.LogFilename, follower.Config{
		Whence: io.SeekEnd,
		Offset: 0,
		Reopen: true,
	})
	if err != nil {
		log.Fatalf("regexp: %s", err)
	}

	regex := `^(\S+) (\S+) (\S+) \[([\w:/]+\s[+\-]\d{4})\] "(\S+) (\S+)\s*(\S+)?\s*" (\d{3}) (\S+)`
	re1, err := regexp.Compile(regex)
	if err != nil {
		log.Fatalf("regexp: %s", err)
	}

	for line := range m.follower.Lines() {
		result := re1.FindStringSubmatch(line.String())
		//TODO: improve this with the regex.
		section := "/" + strings.Split(result[6], "/")[1]

		m.incrementMap(m.hitMap, section, 1)

		bytes, err := strconv.ParseUint(result[9], 10, 64)
		if err != nil {
			log.Fatalf("Parse error: %s", err)
		} else {
			m.incrementMap(m.bytesMap, section, bytes)
		}

		m.incrementMap(m.returnCodeMap, result[8], 1)
	}
}

func (m *HttpLogMonitoring) incrementMap(concurrentMap cmap.ConcurrentMap, key string, value uint64) {
	// collecting sections hits
	if val, ok := concurrentMap.Get(key); ok {
		concurrentMap.Set(key, uint64(val.(uint64)+value))
	} else {
		concurrentMap.Set(key, uint64(value))
	}
}

func (m *HttpLogMonitoring) scrap() {
	var totalHits uint64
	var totalBytes uint64
	var sections []Section

	for item := range m.hitMap.IterBuffered() {

		//section hits
		var hits uint64
		if oldHits, ok := m.oldHitMap.Get(item.Key); ok {
			hits = item.Val.(uint64) - oldHits.(uint64)
		} else {
			hits = item.Val.(uint64)
		}

		totalHits = totalHits + hits
		m.oldHitMap.Set(item.Key, item.Val.(uint64))

		sections = append(sections, Section{Name: item.Key,
			Hits: hits})

		var bytes uint64

		actualBytes, _ := m.bytesMap.Get(item.Key)
		if oldBytes, ok := m.oldBytesMap.Get(item.Key); ok {

			bytes = actualBytes.(uint64) - oldBytes.(uint64)
		} else {
			bytes = actualBytes.(uint64)
		}

		totalBytes = totalBytes + bytes
		m.oldBytesMap.Set(item.Key, actualBytes.(uint64))
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Hits < sections[j].Hits
	})

	//top X

	var topSections = 5
	if len(sections) < topSections {
		topSections = len(sections)
	}

	metrics := Metrics{TotalHits: totalHits,
		TotalBytes:      totalBytes,
		TopSectionsHits: sections[:topSections]}

	m.CollectionChannel <- metrics
}

func (m *HttpLogMonitoring) alert(threshold float64) {

	var totalHits uint64
	var alertAvg float64

	for item := range m.hitMap.IterBuffered() {
		totalHits = totalHits + item.Val.(uint64)
	}

	if m.previousTotalHits > 0 {
		alertAvg = float64(totalHits-m.previousTotalHits) / float64(120)
	} else {
		alertAvg = float64(totalHits) / float64(120)
	}

	if !m.alertTriggered && alertAvg >= threshold {

		m.AlertChannel <- fmt.Sprintf("High traffic generated an alert - hits = %.2f, triggered at %s", alertAvg, time.Unix(time.Now().Unix(), 0))
		m.alertTriggered = true
	}

	if m.alertTriggered && alertAvg < threshold {
		m.AlertChannel <- fmt.Sprintf("Alert recovered at %s", time.Unix(time.Now().Unix(), 0))
		m.alertTriggered = false
	}

	m.previousTotalHits = totalHits
}
