package runtime

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
)

//Settings runtime require settings
type Settings struct {
	GoogleProjectID         string
	BucketName              string
	Port                    int
	GracefulShutdownTimeout time.Duration
	SsoURLKey               string
}

//MustValidate fail if we can't validate
func (s *Settings) MustValidate() {
	if s.GoogleProjectID == "" {
		dieFromMissingVar(projectID)
	}

	if s.BucketName == "" {
		dieFromMissingVar(bucketName)
	}

	if s.SsoURLKey == "" {
		dieFromMissingVar(ssoKeyURL)
	}

}

func dieFromMissingVar(varName string) {
	panic(fmt.Sprintf("You must set the env variable '%s'", varName))
}

//projectID the env var for project it
const projectID = "PROJECTID"

//bucketName the env var for bucket name
const bucketName = "BUCKET_NAME"

//port The port to run on
const port = "PORT"

//the key to the sso url
const ssoKeyURL = "SSO_KEY_URL"

//LoadSettingsFromSystem load the settings from the env vars
func LoadSettingsFromSystem() *Settings {
	v := viper.New()
	v.AutomaticEnv()

	v.SetDefault(port, "5280")

	settings := &Settings{
		GoogleProjectID: v.GetString(projectID),
		BucketName:      v.GetString(bucketName),
		SsoURLKey:       v.GetString(ssoKeyURL),
		Port:            v.GetInt(port),
	}

	log.Printf("Settings are %+v", settings)

	return settings
}
