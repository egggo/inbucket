/*
	The web package contains all the code to provide Inbucket's web GUI
*/
package web

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/egggo/inbucket/config"
	"github.com/egggo/inbucket/log"
	"github.com/egggo/inbucket/smtpd"
	"github.com/goods/httpbuf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type handler func(http.ResponseWriter, *http.Request, *Context) error

var webConfig config.WebConfig
var DataStore smtpd.DataStore
var Router *mux.Router
var listener net.Listener
var sessionStore sessions.Store
var shutdown bool

// Initialize sets up things for unit tests or the Start() method
func Initialize(cfg config.WebConfig, ds smtpd.DataStore) {
	webConfig = cfg
	setupRoutes(cfg)

	// NewContext() will use this DataStore for the web handlers
	DataStore = ds

	// TODO Make configurable
	sessionStore = sessions.NewCookieStore([]byte("something-very-secret"))
}

func setupRoutes(cfg config.WebConfig) {
	log.LogInfo("Theme templates mapped to '%v'", cfg.TemplateDir)
	log.LogInfo("Theme static content mapped to '%v'", cfg.PublicDir)

	r := mux.NewRouter()
	// Static content
	r.PathPrefix("/public/").Handler(http.StripPrefix("/public/",
		http.FileServer(http.Dir(cfg.PublicDir))))

	// Root
	r.Path("/").Handler(handler(RootIndex)).Name("RootIndex").Methods("GET")
	r.Path("/status").Handler(handler(RootStatus)).Name("RootStatus").Methods("GET")
	r.Path("/link/{name}/{id}").Handler(handler(MailboxLink)).Name("MailboxLink").Methods("GET")
	r.Path("/mailbox").Handler(handler(MailboxIndex)).Name("MailboxIndex").Methods("GET")
	r.Path("/mailbox/{name}").Handler(handler(MailboxList)).Name("MailboxList").Methods("GET")
	r.Path("/mailbox/{name}").Handler(handler(MailboxPurge)).Name("MailboxPurge").Methods("DELETE")
	r.Path("/mailbox/{name}/{id}").Handler(handler(MailboxShow)).Name("MailboxShow").Methods("GET")
	r.Path("/mailbox/{name}/{id}/html").Handler(handler(MailboxHtml)).Name("MailboxHtml").Methods("GET")
	r.Path("/mailbox/{name}/{id}/source").Handler(handler(MailboxSource)).Name("MailboxSource").Methods("GET")
	r.Path("/mailbox/{name}/{id}").Handler(handler(MailboxDelete)).Name("MailboxDelete").Methods("DELETE")
	r.Path("/mailbox/dattach/{name}/{id}/{num}/{file}").Handler(handler(MailboxDownloadAttach)).Name("MailboxDownloadAttach").Methods("GET")
	r.Path("/mailbox/vattach/{name}/{id}/{num}/{file}").Handler(handler(MailboxViewAttach)).Name("MailboxViewAttach").Methods("GET")

	// Register w/ HTTP
	Router = r
	http.Handle("/", Router)
}

// Start() the web server
func Start() {
	addr := fmt.Sprintf("%v:%v", webConfig.Ip4address, webConfig.Ip4port)
	server := &http.Server{
		Addr:         addr,
		Handler:      nil,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	// We don't use ListenAndServe because it lacks a way to close the listener
	log.LogInfo("HTTP listening on TCP4 %v", addr)
	var err error
	listener, err = net.Listen("tcp", addr)
	if err != nil {
		log.LogError("HTTP failed to start TCP4 listener: %v", err)
		// TODO More graceful early-shutdown procedure
		panic(err)
	}

	err = server.Serve(listener)
	if shutdown {
		log.LogTrace("HTTP server shutting down on request")
	} else if err != nil {
		log.LogError("HTTP server failed: %v", err)
	}
}

func Stop() {
	log.LogTrace("HTTP shutdown requested")
	shutdown = true
	if listener != nil {
		listener.Close()
	} else {
		log.LogError("HTTP listener was nil during shutdown")
	}
}

// ServeHTTP builds the context and passes onto the real handler
func (h handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Create the context
	ctx, err := NewContext(req)
	if err != nil {
		log.LogError("Failed to create context: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer ctx.Close()

	// Run the handler, grab the error, and report it
	buf := new(httpbuf.Buffer)
	log.LogTrace("Web: %v %v %v %v", req.RemoteAddr, req.Proto, req.Method, req.RequestURI)
	err = h(buf, req, ctx)
	if err != nil {
		log.LogError("Error handling %v: %v", req.RequestURI, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save the session
	if err = ctx.Session.Save(req, buf); err != nil {
		log.LogError("Failed to save session: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Apply the buffered response to the writer
	buf.Apply(w)
}
