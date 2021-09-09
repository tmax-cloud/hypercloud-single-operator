## CLI를 통한 claim류 patch 테스트
1. sample-nsc 생성, hypercloud5-admin sa에 권한 부여
    ```
    $ kubectl apply -f CLI-patch-test.yaml
    ```

2. 실행  
* 2-1. ClusterScope인 경우
    ```
    $ ./CLI-patch-test.sh {CRD 이름} {리소스 이름}
    ```
    예시
    ```
    $ ./CLI-patch-test.sh namespaceclaims swlee-test-nsc
    ```

* 2-2. NamespaceScope인 경우
    ```
    $ ./CLI-patch-test.sh {CRD 이름} {리소스 이름} {네임스페이스}
    ```
    예시
    ```
    $ ./CLI-patch-test.sh resourcequotaclaims swlee-test-nsc swlee-test
    ```
