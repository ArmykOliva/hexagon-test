package main

import "github.com/cookiengineer/gooey/bindings/animations"
import "github.com/cookiengineer/gooey/bindings/canvas2d"
import "github.com/cookiengineer/gooey/bindings/console"
import "github.com/cookiengineer/gooey/bindings/dom"
import "github.com/cookiengineer/gooey/bindings/fetch"
import "battlemap/hexgrid"
import "battlemap/structs"
import "encoding/json"
import "time"

func main() {

	element := dom.Document.QuerySelector("canvas")
	canvas  := canvas2d.ToCanvas(element)

	hexagons := []hexgrid.Hexagon{
		hexgrid.NewHexagon( 0,  0,  0),
		hexgrid.NewHexagon( 1, -1,  0),
		hexgrid.NewHexagon(-1,  1,  0),
		hexgrid.NewHexagon( 0,  1, -1),
		hexgrid.NewHexagon( 0, -1,  1),
		hexgrid.NewHexagon( 1,  0, -1),
		hexgrid.NewHexagon(-1,  0,  1),
	}
	grid := hexgrid.NewMap(1024, 640, 64)

	for _, hexagon := range hexagons {
		grid.Add(&hexagon)
	}

	renderer := hexgrid.NewRenderer(canvas, &grid)


	// TODO Kristof fix this shit

	response, err := fetch.Fetch("http://localhost:3000/api/systems", &fetch.Request{
		Method: fetch.MethodGet,
		Mode:   fetch.ModeSameOrigin,
	})

	if err == nil {

		var system_names []string

		json.Unmarshal(response.Body, &system_names)

		for _, name := range system_names {

			response2, err2 := fetch.Fetch("http://localhost:3000/api/systems/" + name, &fetch.Request{
				Method: fetch.MethodGet,
				Mode:   fetch.ModeSameOrigin,
			})

			if err2 == nil {

				var system structs.System

				json.Unmarshal(response2.Body, &system)

				console.Log(system)

			}

		}

	}

	animations.RequestAnimationFrame(func(timestamp float64) {

		context := renderer.Canvas.GetContext()

		context.BeginPath()
		context.SetFillStyleColor("#ff0000")
		context.FillRect(10, 10, 20, 20)
		context.ClosePath()

		context.BeginPath()
		context.SetStrokeStyleColor("#00ff00")
		context.MoveTo(30, 20)
		context.BezierCurveTo(
			120,  30,
			 30, 120,
			120, 120,
		)
		context.Stroke()
		context.ClosePath()

		context.BeginPath()
		context.Rect(120, 110, 20, 20)
		context.SetStrokeStyleColor("#0000ff")
		context.Stroke()
		context.ClosePath()

		// Render Hexagons on top
		renderer.Render()

	})

	canvas.Element.AddEventListener("mousemove", dom.ToEventListener(func(event *dom.Event) {

		console.Log(event)

	}))

	for true {

		// Do Nothing
		time.Sleep(100 * time.Millisecond)

	}

}
