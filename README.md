# haystack

A bundle storage system for EdgeX bundles

View the [api here](swagger.yaml)


### Glide
still here is what we need to do to have glide working correctly:
+ Make sure `$GOPATH/pkg` `$GOPATH/bin` dirs are empty
+ Each project should have glide.yaml with version

```
- package: github.com/gorilla/mux
  version: v1.3.0
- package: github.com/spf13/viper
  version: 5ed0fc31f7f453625df314d8e66b9791e8d13003
```

+ All projects we have should have same versions reference
+ glide.lock file must be checked in once and updated only on new dependency version update
+ you never do `glide up` anymore always do `glide i`. Do `glide up` only when there is new dependency version update
+ All CI jobs do only `glide install`

### Run local tests
+ `cp tools/env_sample.sh build/env.sh`
+ Set all the variable in `build/env.sh`
+ Drop valid Google Cloud JSON key as `build/svc.json` see the doc on generating one here https://cloud.google.com/vision/docs/common/auth#set_up_a_service_account
+ Run `tools/buildwithcoverage.sh`

## Run application locally
+ `cp tools/env_sample.sh build/env.sh`
+ Set all the variable in `build/env.sh`
+ Drop valid Google Cloud JSON key as `build/svc.json` see the doc on generating one here https://cloud.google.com/vision/docs/common/auth#set_up_a_service_account
+ Run `tools/buildwithcoverage.sh`
