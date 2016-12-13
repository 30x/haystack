#!/bin/sh

if [ ! -f ~/.config/gcloud/application_default_credentials.json ]; then
    echo "Expected the gcloud credentials in ~/.config/gcloud/application_default_credentials.json.  Ensure you log in with https://cloud.google.com/sdk/docs/authorizing"
    exit 1
fi

cp ~/.config/gcloud/application_default_credentials.json .config/gcloud/application_default_credentials.json

if [ "$PROJECTID" == "" ];then
    echo "You must set the PROJECTID environment variable"
fi

cat <<EOF >> buildtime/env.sh
#!/bin/bash
export PROJECTID="$PROJECTID"
EOF