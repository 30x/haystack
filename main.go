package main

import (
	"log"

	"github.com/30x/haystack/api"
	"github.com/30x/haystack/oauth2"
	"github.com/30x/haystack/runtime"
	"github.com/30x/haystack/storage"
)

func main() {

	settings := runtime.LoadSettingsFromSystem()

	settings.MustValidate()

	oAuthService := oauth2.CreateApigeeOAuth(settings.SsoUrlKey)

	storage, err := storage.CreateGCloudStorage(settings.GoogleProjectID, settings.BucketName)

	if err != nil {
		log.Fatal(err)
	}

	routes := api.CreateRoutes(storage, oAuthService)

	runtime := runtime.CreateRuntime(routes, settings.Port, settings.GracefulShutdownTimeout)

	err = runtime.Start()

	if err != nil {
		log.Fatal(err)
	}
}
