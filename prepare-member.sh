#/bin/bash

make switch-to-host
echo "getting host sa token"
HOST_SA_TOKEN=`oc serviceaccounts get-token saas-next`
make set-profile-to-member
make create-resources
echo "creating host sa token in member"
oc create secret generic host-sa-saas-next-token --from-literal=token=${HOST_SA_TOKEN}
