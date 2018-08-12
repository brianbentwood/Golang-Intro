package route

import (
	"net/http"

	"../controller"
	"../shared/session"
	"./middleware/acl"
	hr "./middleware/httprouterwrapper"
	"./middleware/logrequest"
	"./middleware/pprofhandler"

	"github.com/gorilla/context"
	"github.com/josephspurrier/csrfbanana"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

// Load returns the routes and middleware
func Load() http.Handler {
	return middleware(routes())
}

// LoadHTTPS returns the HTTP routes and middleware
func LoadHTTPS() http.Handler {
	return middleware(routes())
}

// LoadHTTP returns the HTTPS routes and middleware
func LoadHTTP() http.Handler {
	return middleware(routes())

	// Uncomment this and comment out the line above to always redirect to HTTPS
	//return http.HandlerFunc(redirectToHTTPS)
}

// Optional method to make it easy to redirect from HTTP to HTTPS
func redirectToHTTPS(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "https://"+req.Host, http.StatusMovedPermanently)
}

// *****************************************************************************
// Routes
// *****************************************************************************

func routes() *httprouter.Router {
	r := httprouter.New()

	// Set 404 handler
	r.NotFound = alice.
		New().
		ThenFunc(controller.Error404)

	// Serve static files, no directory browsing
	//r.GET("/static/*filepath", hr.Handler(alice.
	//	New().
	//	ThenFunc(controller.Static)))

	// Home page
	r.GET("/", hr.Handler(alice.
		New().
		ThenFunc(controller.IndexGET)))

	// About
	r.GET("/about", hr.Handler(alice.
		New().
		ThenFunc(controller.AboutGET)))

	// Search
	//r.GET("/search", hr.Handler(alice.
	//	New(acl.DisallowAnon).
	//	ThenFunc(controller.SearchGET)))
	//r.POST("/search", hr.Handler(alice.
	//	New(acl.DisallowAnon).
	//	ThenFunc(controller.SearchPOST)))

	r.GET("/search", hr.Handler(alice.
		New().
		ThenFunc(controller.SearchGET)))
	r.POST("/search", hr.Handler(alice.
		New().
		ThenFunc(controller.SearchPOST)))

	/*
		r.GET("/search/create", hr.Handler(alice.
			New(acl.DisallowAnon).
			ThenFunc(controller.SearchCreateGET)))
		r.POST("/search/create", hr.Handler(alice.
			New(acl.DisallowAnon).
			ThenFunc(controller.SearchCreatePOST)))
		r.GET("/search/update/:id", hr.Handler(alice.
			New(acl.DisallowAnon).
			ThenFunc(controller.SearchUpdateGET)))
		r.POST("/search/update/:id", hr.Handler(alice.
			New(acl.DisallowAnon).
			ThenFunc(controller.SearchUpdatePOST)))
		r.GET("/search/delete/:id", hr.Handler(alice.
			New(acl.DisallowAnon).
			ThenFunc(controller.SearchDeleteGET)))
	*/
	// Enable Pprof
	r.GET("/debug/pprof/*pprof", hr.Handler(alice.
		New(acl.DisallowAnon).
		ThenFunc(pprofhandler.Handler)))

	return r
}

// *****************************************************************************
// Middleware
// *****************************************************************************

func middleware(h http.Handler) http.Handler {
	// Prevents CSRF and Double Submits
	cs := csrfbanana.New(h, session.Store, session.Name)
	cs.FailureHandler(http.HandlerFunc(controller.InvalidToken))
	cs.ClearAfterUsage(true)
	cs.ExcludeRegexPaths([]string{"/static(.*)"})
	csrfbanana.TokenLength = 32
	csrfbanana.TokenName = "token"
	csrfbanana.SingleToken = false
	h = cs

	// Log every request
	h = logrequest.Handler(h)

	// Clear handler for Gorilla Context
	h = context.ClearHandler(h)

	return h
}
