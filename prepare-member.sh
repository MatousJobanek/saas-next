#/bin/bash

make switch-to-host
echo "getting host sa token"
HOST_SA_SECRET=`oc get sa saas-next -o json | jq -r .secrets[].name | grep token`
HOST_SA_TOKEN=`oc get secret ${HOST_SA_SECRET} -o json | jq -r '.data["token"]' | base64 -d`
HOST_SA_CA_CRT=`oc get secret ${HOST_SA_SECRET} -o json | jq -r '.data["ca.crt"]' | base64 -d`
make set-profile-to-member
make create-resources
echo "creating host sa token in member"

oc create secret generic host-sa-saas-next-token --from-literal=token="${HOST_SA_TOKEN}" --from-literal=ca.crt="${HOST_SA_CA_CRT}"
