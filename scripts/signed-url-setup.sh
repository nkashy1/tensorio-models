#!/usr/bin/env sh

# Accepts a GCP service account e-mail and produces a PEM file for that service account at the
# given path
# A GCP service account e-mail looks like this:
# <SERVICE_ACCOUNT_NAME>@<GCP_PROJECT>.iam.gserviceaccount.com
#
# Requirements:
# + Correctly configured gcloud command on $PATH (for more information: https://cloud.google.com/sdk/)
# + openssl on $PATH (for more information: https://wiki.openssl.org/index.php/Command_Line_Utilities)

SERVICE_ACCOUNT=$1
if [ -z $SERVICE_ACCOUNT ]; then
    echo "ERROR: Please pass GCP service account e-mail address as first argument"
    exit 1
fi

SECRETS_DIR=$2
if [ -z $SECRETS_DIR ]; then
    SECRETS_DIR=$(mktemp -d)
    echo "WARNING: Second argument not provided -- secrets being stored in $SECRETS_DIR" 1>&2
fi

TIMESTAMP=$(date -u +%Y%m%d-%H%M%S)
P12_FILENAME="$SECRETS_DIR/sacred-${SERVICE_ACCOUNT}-${TIMESTAMP}.p12"
PEM_FILENAME="$SECRETS_DIR/sacred-${SERVICE_ACCOUNT}-${TIMESTAMP}.pem"

echo "Creating P12 key: $P12_FILENAME"
gcloud iam service-accounts keys create "$P12_FILENAME" \
    --iam-account=$SERVICE_ACCOUNT \
    --key-file-type=p12

echo "Creating PEM file: $PEM_FILENAME"
openssl pkcs12 -in $P12_FILENAME -passin pass:notasecret -out $PEM_FILENAME -nodes

echo "Done!"
