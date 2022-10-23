export APPLICATION_NAME="cssd-sample-app"
export APPLICATION_CLIENT_ID=$(az ad sp list --display-name ${APPLICATION_NAME} --query '[0].appId' -otsv)
export SERVICE_ACCOUNT_NAME="cssd-sample-sa"
export SERVICE_ACCOUNT_NAMESPACE="cssd-azure-test"
export APPLICATION_OBJECT_ID="$(az ad app show --id ${APPLICATION_CLIENT_ID} --query id -otsv)"