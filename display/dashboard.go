package display

import (
	"logmonitor-homework/monitoring"

	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type Dashboard struct {
	CollectionInterval int
	CollectionChannel  chan monitoring.Metrics
	AlertInterval      int
	AlertThreshold     float64
	AlertChannel       chan string
}

func (d *Dashboard) DisplayStart() {

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	p := widgets.NewParagraph()
	p.Text = "HTTP Log monitor via access logs ( Press q to exit )"
	p.SetRect(0, 0, 70, 3)
	p.Border = false
	p.WrapText = false

	slth := widgets.NewSparkline()
	slth.Title = "Total Hits:"
	slth.LineColor = ui.ColorCyan
	slth.TitleStyle.Fg = ui.ColorWhite

	sltb := widgets.NewSparkline()
	sltb.Title = "Total Bytes:"
	sltb.TitleStyle.Fg = ui.ColorWhite
	sltb.LineColor = ui.ColorRed

	slg := widgets.NewSparklineGroup(slth, sltb)
	slg.Title = "General Metrics - 10 seconds aggregation"
	slg.SetRect(0, 3, 70, 14)

	l := widgets.NewList()
	l.Title = "Top 5 Section Hits "
	//l.Rows = pagesList
	l.TextStyle = ui.NewStyle(ui.ColorYellow)
	l.WrapText = false
	l.SetRect(71, 0, 95, 14)

	la := widgets.NewList()
	la.Title = "Alerts "
	//l.Rows = pagesList
	la.TextStyle = ui.NewStyle(ui.ColorYellow)
	la.WrapText = false
	la.SetRect(0, 15, 95, 25)

	draw := func() {

		ui.Render(p, l, slg, la)
	}

	go func() {
		for {
			select {
			case metrics := <-d.CollectionChannel:
				var sectionList []string
				//fmt.Printf("DRAW = %d", metrics.TotalHits)

				for i := len(metrics.TopSectionsHits) - 1; i >= 0; i-- {
					sectionList = append(sectionList, fmt.Sprintf("%s Hits : %d ", metrics.TopSectionsHits[i].Name, metrics.TopSectionsHits[i].Hits))
				}

				// total hits
				slth.Title = fmt.Sprintf("Total Hits: %d", metrics.TotalHits)
				slth.Data = append(slth.Data, float64(metrics.TotalHits))

				// total bytes
				sltb.Title = fmt.Sprintf("Total Bytes: %d", metrics.TotalBytes)
				sltb.Data = append(sltb.Data, float64(metrics.TotalBytes))

				l.Rows = sectionList

			case alert := <-d.AlertChannel:
				la.Rows = append([]string{alert}, la.Rows...)
			}
		}
	}()

	draw()

	uiEvents := ui.PollEvents()
	//ticker := time.NewTicker(time.Second).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		default:
			time.Sleep(time.Second)
			draw()
		}
	}

}
