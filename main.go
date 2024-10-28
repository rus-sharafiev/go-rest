package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/rus-sharafiev/go-rest-auth/auth"
	common "github.com/rus-sharafiev/go-rest-common"
	"github.com/rus-sharafiev/go-rest-common/db"
	"github.com/rus-sharafiev/go-rest-common/formdata"
	"github.com/rus-sharafiev/go-rest-common/spa"
	"github.com/rus-sharafiev/go-rest-common/uploads"
	"github.com/rus-sharafiev/go-rest/chat"
	"github.com/rus-sharafiev/go-rest/images"
	"github.com/rus-sharafiev/go-rest/user"
)

//go:embed config.json
var config []byte

func main() {
	// Load app config
	if err := json.Unmarshal(config, &common.Config); err != nil {
		log.Fatalf("\n\x1b[31m Error parsing the config file: %v\x1b[0m\n", err)
	} else if common.Config.IsNotValid() {
		log.Fatalf("\x1b[31mThe config file is missing required fields \x1b[0m\n\n")
	}

	// Connect to the database and create HTTP request multiplexer
	db.Connect(*common.Config.Db)
	mux := http.NewServeMux()

	// API ----------------------------------------------------------------------------

	mux.Handle("/api/auth/", auth.Controller)
	mux.Handle("/api/users/", user.Controller)
	mux.Handle("/api/chats/", chat.Controller)

	// --------------------------------------------------------------------------------

	// Specific pathes
	mux.Handle("/images/", images.Controller)

	// Static files
	mux.Handle(*common.Config.UploadPath, uploads.Handler)
	mux.Handle("/", spa.Handler)

	// Middleware
	handler := formdata.Interceptor(mux)
	handler = auth.Guard(handler)
	handler = cors.New(cors.Options{
		AllowedOrigins:   []string{"http://192.168.190.9:5555", "http://192.168.190.9:8000", "http://localhost:8000"},
		AllowedHeaders:   []string{"Content-Type", "Fingerprint", "Authorization"},
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowCredentials: true,
	}).Handler(handler)

	fmt.Printf("\n\x1b[32mServer is running on port %v\x1b[0m\n\n", *common.Config.Port)
	log.Fatal(http.ListenAndServe(":"+*common.Config.Port, handler))
}
