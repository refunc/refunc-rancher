refunc-rancher
========

rancher integration

## Building

`make`


## Running

`./bin/refunc-rancher`

## Dev on testing env

1. Install using `kubectl`
    ```shell
    kubectl create -f ./k8s
    ```

1. Login to rancher, build your own porxy URL like the following
    ```
    https://rancher.<your-domain>.com/k8s/clusters/<cluster-id-where-refunc-deployed>/api/v1/namespaces/refunc/services/http:refunc-rancher:80/proxy/
    ```

## Screenshots

![functions.png](https://user-images.githubusercontent.com/354668/44694551-b13f3900-aaa0-11e8-8a9a-a19d562ec8d1.png "Functions page")

![funcinst.png](https://user-images.githubusercontent.com/354668/44694576-cd42da80-aaa0-11e8-87b8-dedce53e420f.png "Funcinst page")

## License
Copyright (c) 2018 [refunc.io](http://refunc.io)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
