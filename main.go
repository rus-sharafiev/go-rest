package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/cors"
	"github.com/rus-sharafiev/go-rest/common/auth"
	"github.com/rus-sharafiev/go-rest/common/formdata"
	"github.com/rus-sharafiev/go-rest/common/spa"
	"github.com/rus-sharafiev/go-rest/common/uploads"
	"github.com/rus-sharafiev/go-rest/images"
	"github.com/rus-sharafiev/go-rest/user"
)

var (
	staticDir  = ".static"
	uploadPath = "/upload/"
	uploadDir  = filepath.Join(staticDir, "upload")
)

func main() {
	loadEnv()

	port := flag.String("port", "", "PORT to run http handler")
	flag.Parse()

	if len(*port) == 0 {
		fmt.Println("provide port number")
		flag.Usage()
		os.Exit(0)
	}

	router := http.NewServeMux()

	// API
	router.Handle("/api/auth/", auth.Controller)
	router.Handle("/api/users/", user.Controller)

	// Handle and serve images
	router.Handle("/images/", &images.Controller{
		UploadDir: filepath.Join(staticDir, "images"),
	})

	// Serve uploads made by form data interceptor
	router.Handle(uploadPath, &uploads.Server{
		Dir: &staticDir,
	})

	// Serve static files and SPA
	router.Handle("/", &spa.Handler{
		Static: &staticDir,
	})

	handler := formdata.Interceptor{
		UploadDir:  &uploadDir,
		UploadPath: &uploadPath,
	}.Handler(router)
	handler = auth.Guard(handler)
	handler = cors.New(cors.Options{
		AllowedOrigins:   []string{"http://192.168.190.9:5555", "http://192.168.190.9:8000", "http://localhost:8000"},
		AllowedHeaders:   []string{"Content-Type", "Fingerprint", "Authorization"},
		AllowCredentials: true,
	}).Handler(handler)

	fmt.Println("server is running on port " + *port)
	log.Fatal(http.ListenAndServe(":"+*port, handler))
}
