package runtime

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

//Settings runtime require settings
type Settings struct {
	GoogleProjectID         string
	BucketName              string
	Port                    int
	GracefulShutdownTimeout time.Duration
	SsoUrlKey               string
}

//MustValidate fail if we can't validate
func (s *Settings) MustValidate() {
	if s.GoogleProjectID == "" {
		dieFromMissingVar(projectID)
	}

	if s.BucketName == "" {
		dieFromMissingVar(bucketName)
	}

	if s.SsoUrlKey == "" {
		dieFromMissingVar(ssoKeyURl)
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
const ssoKeyURl = "SSO_KEY_URL"

//LoadSettingsFromSystem load the settings from the env vars
func LoadSettingsFromSystem() *Settings {
	v := viper.New()
	v.AutomaticEnv()

	v.SetDefault(port, "5280")

	settings := &Settings{
		GoogleProjectID: v.GetString(projectID),
		BucketName:      v.GetString(bucketName),
		SsoUrlKey:       v.GetString(ssoKeyURl),
		Port:            v.GetInt(port),
	}

	return settings
}
