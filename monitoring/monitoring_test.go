package monitoring

import (
	"github.com/orcaman/concurrent-map"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestHttpLogMonitoring_Monitor(t *testing.T) {

	emptyFile, err := os.Create("/tmp/unit-test.log")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(emptyFile)
	emptyFile.Close()

	type fields struct {
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
	}

	type args struct {
		line string
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		//{name:"file not exists Test Tail",
		//	fields:fields{LogFilename:"/tmp/unit-test.log"}},
		{name: "Test Tail ragex",
			fields: fields{LogFilename: "/tmp/unit-test.log"},
			args:   args{line: "127.0.0.1 - james [09/May/2018:16:00:39 +0000] \"GET /pages HTTP/1.0\" 200 123\n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HttpLogMonitoring{
				CollectionInterval: tt.fields.CollectionInterval,
				CollectionChannel:  tt.fields.CollectionChannel,
				AlertInterval:      tt.fields.AlertInterval,
				AlertThreshold:     tt.fields.AlertThreshold,
				AlertChannel:       tt.fields.AlertChannel,
				LogFilename:        tt.fields.LogFilename,
				returnCodeMap:      tt.fields.returnCodeMap,
				hitMap:             tt.fields.hitMap,
				bytesMap:           tt.fields.bytesMap,
				oldBytesMap:        tt.fields.oldBytesMap,
				oldHitMap:          tt.fields.oldHitMap,
				previousTotalHits:  tt.fields.previousTotalHits,
				alertTriggered:     tt.fields.alertTriggered,
			}

			go m.Monitor()
			go func() {

				// If the file doesn't exist, create it, or append to the file
				f, err := os.OpenFile("/tmp/unit-test.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					log.Fatal(err)
				}

				for {

					if _, err := f.Write([]byte(tt.args.line)); err != nil {
						log.Fatal(err)
					}

				}
				if err := f.Close(); err != nil {
					log.Fatal(err)
				}

				time.Sleep(time.Second)
				actual, _ := m.hitMap.Get("/pages")
				assert.EqualValues(t, 1, actual)

			}()

			time.Sleep(time.Second * 5)
			actual, _ := m.hitMap.Get("/pages")

			assert.NotNil(t, actual.(uint64))

		})
	}
}

func TestHttpLogMonitoring_scrap(t *testing.T) {

	hitmap := cmap.New()
	hitmap.Set("/pages", uint64(2400))

	bytesMap := cmap.New()
	bytesMap.Set("/pages", uint64(1000))

	oldHitMap := cmap.New()
	oldBytesMap := cmap.New()

	collectionChannel := make(chan Metrics)

	type fields struct {
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
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{name: "Alert_Triggered_after_first_scrap",
			fields: fields{
				hitMap:            hitmap,
				CollectionChannel: collectionChannel,
				oldHitMap:         oldHitMap,
				oldBytesMap:       oldBytesMap,
				bytesMap:          bytesMap,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HttpLogMonitoring{
				CollectionInterval: tt.fields.CollectionInterval,
				CollectionChannel:  tt.fields.CollectionChannel,
				AlertInterval:      tt.fields.AlertInterval,
				AlertThreshold:     tt.fields.AlertThreshold,
				AlertChannel:       tt.fields.AlertChannel,
				LogFilename:        tt.fields.LogFilename,
				returnCodeMap:      tt.fields.returnCodeMap,
				hitMap:             tt.fields.hitMap,
				bytesMap:           tt.fields.bytesMap,
				oldBytesMap:        tt.fields.oldBytesMap,
				oldHitMap:          tt.fields.oldHitMap,
				previousTotalHits:  tt.fields.previousTotalHits,
				alertTriggered:     tt.fields.alertTriggered,
			}
			go m.scrap()

			metrics := <-m.CollectionChannel
			assert.Equal(t, uint64(2400), metrics.TotalHits)
		})
	}
}

func TestHttpLogMonitoring_alert(t *testing.T) {

	hitmap := cmap.New()
	hitmap.Set("/pages", uint64(2400))

	previousTotalHits := uint64(1200)

	alertChannel := make(chan string)

	type fields struct {
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
	}
	type args struct {
		threshold float64
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		expected string
	}{
		{name: "Alert_Triggered_first_scrap",
			fields: fields{
				hitMap:       hitmap,
				AlertChannel: alertChannel,
			},
			args:     args{threshold: 10},
			expected: "hits = 20.00"},
		{name: "Alert_Triggered_after_first_scrap",
			fields: fields{
				hitMap:            hitmap,
				AlertChannel:      alertChannel,
				previousTotalHits: previousTotalHits,
			},
			args:     args{threshold: 10},
			expected: "hits = 10.00"},
		{name: "Alert_recovered",
			fields: fields{
				hitMap:            hitmap,
				AlertChannel:      alertChannel,
				previousTotalHits: uint64(2300),
				alertTriggered:    true,
			},
			args:     args{threshold: 10},
			expected: "Alert recovered at"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &HttpLogMonitoring{
				CollectionInterval: tt.fields.CollectionInterval,
				CollectionChannel:  tt.fields.CollectionChannel,
				AlertInterval:      tt.fields.AlertInterval,
				AlertThreshold:     tt.fields.AlertThreshold,
				AlertChannel:       tt.fields.AlertChannel,
				LogFilename:        tt.fields.LogFilename,
				returnCodeMap:      tt.fields.returnCodeMap,
				hitMap:             tt.fields.hitMap,
				bytesMap:           tt.fields.bytesMap,
				oldBytesMap:        tt.fields.oldBytesMap,
				oldHitMap:          tt.fields.oldHitMap,
				previousTotalHits:  tt.fields.previousTotalHits,
				alertTriggered:     tt.fields.alertTriggered,
			}

			go m.alert(tt.args.threshold)
			result := <-m.AlertChannel
			assert.Contains(t, result, tt.expected)
		})
	}
}
