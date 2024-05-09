// Our empty version of the httpServer for usage with the wasm target
// this way we will not include any of the related code
//go:build !wasm

package main

import (
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func main() {
	Front()

	// this concludes the part which goes into the front-end
	app.RunWhenOnBrowser()

	// I declare this here because of the "logic" to find it in
	// the backend code still is a mindbender for me :)
	ah := &app.Handler{
		Name:               "Go-Nats-App",
		Lang:               "de",
		Author:             "Hans Raaf - METATEXX GmbH",
		Title:              "Go Nats App",
		Description:        "NATS in a PWA",
		Image:              "/web/logo-512.png",
		LoadingLabel:       "Loading...",
		AutoUpdateInterval: 5 * time.Second, // so we do not need to refresh the browser tabs while experimenting
		Icon: app.Icon{
			Default:    "/web/logo-192.png",
			Large:      "/web/logo-512.png",
			AppleTouch: "/web/logo-192.png",
		},
		Styles: []string{
			"/web/index.css",
		},
		CacheableResources: []string{
			"/web/logo.svg",
		},
	}

	Back(ah)
}

type empty struct{ app.Compo }

// Render implements the "source" code our App is showing before WASM is loaded and runs.
// Actually the WASM loader takes over if javascript is available. So this is mainly
// what robots and scrappers get to see.
func (c *empty) Render() app.UI {
	return app.Div().Text("The application is loading...")
}

func Front() {
	// This will skip most of the frontend code in the server binary
	// but keeps the routing intact. It is important that all possible
	// routes that are reachable in the frontend are here. Any route that
	// is not covered results in a (raw) 404 error from the server.

	// For the Skelly demo app, we handle all routes in the frontend itself
	app.RouteWithRegexp("/.*", &empty{})
}

func Back(ah *app.Handler) {
	Create(ah)
	//back.Create(ah)
}
