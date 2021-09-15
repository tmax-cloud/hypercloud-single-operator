#!/bin/bash
# 1. 클러스터 이름을 참조하는 API 서버를 가리킨다.
CLUSTER_NAME=$(kubectl config view -o jsonpath="{.clusters[0].name}")
APISERVER=$(kubectl config view -o jsonpath="{.clusters[?(@.name==\"$CLUSTER_NAME\")].cluster.server}")


# 2. 토큰 값을 얻는다
TOKEN=$(kubectl get secrets -n hypercloud5-system -o jsonpath="{.items[?(@.metadata.annotations['kubernetes\.io/service-account\.name']=='hypercloud5-admin')].data.token}"|base64 --decode)


# 3. TOEKN으로 API 콜
if [ -z "$3" ]
then
    # 3-1. cluster-scope의 리소스인 경우
    # $1={CRD 이름} $2={리소스 이름}
    curl -X PATCH "$APISERVER/apis/claim.tmax.io/v1alpha1/$1/$2/status" -H "Content-Type: application/json-patch+json" -H "Authorization: Bearer $TOKEN" -k -d '[{"op": "replace", "path": "/status/status", "value": "Approved"}]'
else
    # 3-2. namespace-scope의 리소스인 경우
    # $1={CRD 이름} $2={리소스 이름} $3={네임스페이스}
    curl -X PATCH "$APISERVER/apis/claim.tmax.io/v1alpha1/namespaces/$3/$1/$2/status" -H "Content-Type: application/json-patch+json" -H "Authorization: Bearer $TOKEN" -k -d '[{"op": "replace", "path": "/status/status", "value": "Approved"}]'
fi

