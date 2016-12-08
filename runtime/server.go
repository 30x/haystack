package runtime

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

	"github.com/gorilla/handlers"
	"github.com/tylerb/graceful"
)

//Runtime the instance of our server
type Runtime struct {
	server *graceful.Server
	port   int
}

//CreateRuntime Create a new server runtime and return it.  Adds all services
func CreateRuntime(routes *mux.Router, port int, shutdownTimeout time.Duration) *Runtime {

	//TODO refactor and clean this up.  What is a more effective way of organizing this?

	//wire up our middleware
	server := &Runtime{
		port: port,
	}

	//status page

	//now wrap everything with logging and panic recovery
	loggedRouter := handlers.RecoveryHandler()(handlers.CombinedLoggingHandler(os.Stdout, routes))

	address := fmt.Sprintf(":%d", port)

	log.Printf("Starting server at address %s", address)

	//TODO add negroni middleware for securing endpoints that require a google oAuth token
	server.server = &graceful.Server{
		Timeout: shutdownTimeout,
		Server:  &http.Server{Addr: address, Handler: loggedRouter},
		Logger:  graceful.DefaultLogger(),
	}

	//set up our timer in a gofunc in order to shut down after a duration

	return server

}

//Start start the server
func (server *Runtime) Start() error {
	//start listening
	if err := server.server.ListenAndServe(); err != nil {
		if opErr, ok := err.(*net.OpError); !ok || (ok && opErr.Op != "accept") {
			return err
		}
	}

	return nil

}

//Stop stop the server
func (server *Runtime) Stop() {
	server.server.Stop(0 * time.Second)
}

//Port return the port the server is running
func (server *Runtime) Port() int {
	return server.port
}
