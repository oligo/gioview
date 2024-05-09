package main

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/o1egl/govatar"
	"golang.org/x/net/netutil"
)

// AppServer implements the Backend service
type AppServer struct {
	mux *chi.Mux
}

func (srv *AppServer) Mux() *chi.Mux {
	return srv.mux
}

type Muxer interface {
	Mux() *chi.Mux
}

func Create(ah *app.Handler) {
	var srv AppServer

	opts := &server.Options{
		ServerName:     "Your friendly backend",
		Host:           "127.0.0.1",
		Port:           8501,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 4096,
		Accounts: []*server.Account{
			{
				Name: "cluster",
			},
		},
		Websocket: server.WebsocketOpts{
			Host:             "127.0.0.1",
			Port:             8502,
			NoTLS:            true,
			SameOrigin:       false,
			Compression:      false,
			HandshakeTimeout: 5 * time.Second,
		},
		DisableShortFirstPing: true,
	}
	// Initialize new server with options
	ns, err := server.NewServer(opts)
	if err != nil {
		panic(err)
	}

	// Start the server via goroutine
	go ns.Start()

	// Wait for server to be ready for connections
	if !ns.ReadyForConnections(2 * time.Second) {
		panic("could not start the server (is another instance running on the same port?)")
	}

	fmt.Printf("Nats AppServer Name: %q\n", ns.Name())
	fmt.Printf("Nats AppServer Addr: %q\n", ns.Addr())

	// Connect to server
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		panic(err)
	}

	// This is the chat broker
	_, err = nc.Subscribe("chat.say", func(msg *nats.Msg) {
		fmt.Printf("Got: %q\n", msg.Data)
		err = nc.Publish("chat.room", msg.Data)
		if err != nil {
			panic(err)
		}
	})
	if err != nil {
		panic(err)
	}

	_, err = nc.Subscribe("govatar.female", func(msg *nats.Msg) {
		if err != nil {
			panic(err)
		}
		var img image.Image
		// always female and random
		img, err = govatar.Generate(govatar.FEMALE)
		if err != nil {
			panic(err)
		}
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
		if err != nil {
			panic(err)
		}
		err = msg.Respond([]byte("data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())))
		if err != nil {
			log.Printf("Respond error: %s", err)
		}
		//noErr(err)
	})
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	srv.mux = r

	r.Use(middleware.CleanPath)
	r.Use(middleware.Logger)
	compressor := middleware.NewCompressor(flate.DefaultCompression,
		"application/wasm", "text/css", "image/svg+xml" /*, "application/javascript"*/)
	r.Use(compressor.Handler)
	r.Use(middleware.Recoverer)

	listener, err := net.Listen("tcp", "127.0.0.1:8500")
	if err != nil {
		panic(err)
	}

	mainServer := &http.Server{
		Addr: listener.Addr().String(),
	}

	// try to build the final browser location
	var hostURL = url.URL{
		Scheme: "http",
		Host:   mainServer.Addr,
	}

	log.Printf("Serving at %s\n", hostURL.String())

	// This is the frontend handler (go-app) and will "pick up" any route
	// which is not handled by the backend. It then will load the frontend
	// and navigate it to this router.
	srv.mux.Handle("/*", ah)

	// registering our stack as the handler for the http server
	mainServer.Handler = srv.Mux()

	// Creating some graceful shutdown system
	shutdown := make(chan struct{})
	go func() {
		listener = netutil.LimitListener(listener, 100)
		defer func() { _ = listener.Close() }()

		err := mainServer.Serve(listener)
		if err != nil {
			if err == http.ErrServerClosed {
				fmt.Println("AppServer was stopped!")
			} else {
				log.Fatal(err)
			}
		}
		// when run by "mage watch/run" it will break mage
		// before actually exiting the server, which just looks
		// strange but is okay after all
		//time.Sleep(2 * time.Second) // just for testing
		close(shutdown)
	}()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	// Waiting for SIGINT (kill -SIGINT <pid> or Ctrl+c )
	<-stop
	shutdownServer(mainServer, shutdown)
	fmt.Println("Program ended")
}

// shutdownServer stops the server gracefully
func shutdownServer(server *http.Server, shutdown chan struct{}) {
	fmt.Println("\nStopping AppServer gracefully")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("Send shutdown!")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Waiting for server to shutdown!")
	<-shutdown
	fmt.Println("AppServer is shutdown!")
}
