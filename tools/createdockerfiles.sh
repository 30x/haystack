#!/bin/sh

if [ ! -f ~/.config/gcloud/application_default_credentials.json ]; then
    echo "Expected the gcloud credentials in ~/.config/gcloud/application_default_credentials.json.  Ensure you log in with https://cloud.google.com/sdk/docs/authorizing"
    exit 1
fi

#Copy over our creds to the build information
cp ~/.config/gcloud/application_default_credentials.json build/application_default_credentials.json

if [ "$PROJECTID" == "" ];then
    echo "You must set the PROJECTID environment variable"
fi

mkdir -p build

cat <<EOF > build/env.sh
#!/bin/bash
export PROJECTID="$PROJECTID"
EOF