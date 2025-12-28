package router

import (
	"dpv/dpv/src/api"
	"dpv/dpv/src/endpoints/clubs"
	"dpv/dpv/src/endpoints/users"
	"dpv/dpv/src/middleware"
	"dpv/dpv/src/repository/dpv"
	"dpv/dpv/src/repository/graph"
	"dpv/dpv/src/repository/storage"
	"dpv/dpv/src/repository/t"
	"dpv/dpv/src/service/club"
	"dpv/dpv/src/service/user"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

func NewServer(configPath string, test bool) *http.Server {
	attempts := 0
	if !test {
		attempts = 5
	}
	db, config, err := graph.Init(configPath, test)
	for attempt := 0; attempt < attempts; attempt++ {
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * time.Duration(attempt*attempt)) // 0, 1, 4, 9, 16 seconds
			db, config, err = graph.Init(configPath, test)
			continue
		}
		break
	}
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
	}
	dpv.ConfigInstance = config

	r := httprouter.New()
	userService := user.NewService(db)
	userHandler := users.NewHandler(userService)

	st := storage.NewStorage(dpv.ConfigInstance.Storage.DocumentPath)
	clubService := club.NewService(db, st)
	clubHandler := clubs.NewHandler(clubService)

	r.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Access-Control-Request-Method") != "" {
			header := w.Header()
			header.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-altcha-spam-filter")
			header.Set("Access-Control-Allow-Methods", header.Get("Allow"))
			header.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		}

		w.WriteHeader(http.StatusNoContent)
	})

	r.GET("/dpv/version", Version)
	r.POST("/dpv/users", userHandler.Register)
	r.GET("/dpv/users/me", middleware.BasicAuthMiddleware(userHandler.Me, db))
	r.PATCH("/dpv/users/me", middleware.BasicAuthMiddleware(userHandler.UpdateMe, db))

	r.POST("/dpv/users/request-email-validation", middleware.BasicAuthMiddleware(userHandler.RequestEmailValidation, db))
	r.GET("/dpv/users/validate-email", userHandler.ValidateEmail)

	// Add password reset endpoints
	r.POST("/dpv/users/request-password-reset", userHandler.RequestPasswordReset)
	r.GET("/dpv/users/reset-password", userHandler.ShowResetPasswordForm)
	r.POST("/dpv/users/reset-password", userHandler.HandleResetPassword)
	r.PATCH("/dpv/users/:key/roles", userHandler.UpdateRoles)

	r.POST("/dpv/clubs", middleware.BasicAuthMiddleware(clubHandler.Create, db))
	r.GET("/dpv/clubs", middleware.BasicAuthMiddleware(clubHandler.List, db))
	r.GET("/dpv/clubs/:key", middleware.BasicAuthMiddleware(clubHandler.Get, db))
	r.PATCH("/dpv/clubs/:key", middleware.BasicAuthMiddleware(clubHandler.Update, db))
	r.DELETE("/dpv/clubs/:key", middleware.BasicAuthMiddleware(clubHandler.Delete, db))

	r.POST("/dpv/clubs/:key/apply", middleware.BasicAuthMiddleware(clubHandler.Apply, db))
	r.POST("/dpv/clubs/:key/approve", middleware.BasicAuthMiddleware(clubHandler.Approve, db))
	r.POST("/dpv/clubs/:key/deny", middleware.BasicAuthMiddleware(clubHandler.Deny, db))
	r.POST("/dpv/clubs/:key/cancel", middleware.BasicAuthMiddleware(clubHandler.Cancel, db))
	r.POST("/dpv/clubs/:key/documents", middleware.BasicAuthMiddleware(clubHandler.UploadDocument, db))

	r.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		log.Printf("panic: %+v", err)
		api.Error(w, r, t.Errorf("Whoops! It seems we've stumbled upon a glitch here. In the meantime, consider this a chance to take a breather."), http.StatusInternalServerError)
	}
	r.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Error(w, r, t.Errorf("Oops, your %v move is impressive, but this method doesn't match the route's rhythm. Let's stick to the right Parkour technique – we've got OPTIONS waiting for you, not this wild %v dance!", r.Method, r.Method), http.StatusMethodNotAllowed)
	})
	r.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api.Error(w, r, t.Errorf("Oops, you're performing a daring stunt! But this route seems to be off our servers. Maybe let's stick to known paths for now and avoid tumbling into the broken API!"), http.StatusNotFound)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	addr := "localhost:" + port
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func Version(w http.ResponseWriter, r *http.Request, urlParams httprouter.Params) {
	// the only endpoint that does not use JSON-formatted response, i.e. no quotes around version string
	api.Success(w, r, []byte(dpv.ConfigInstance.Settings.Version))
}
