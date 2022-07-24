# secretbox-operator
Secretbox-operator is a k8s operator which generates secret, configmap, deployment and service for a single node secretbox instance.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

### Install controller

```
git clone https://github.com/pavel1337/secretbox-operator
cd secretbox-operator
kubectl apply -f dist/
```

### Create secretbox instance

You can create a sample secretbox instance by running:
```bash
kubectl apply -f - <<EOF
apiVersion: secretbox.ipvl.de/v1
kind: Secretbox
metadata:
  name: sample
spec:
  size: 1
EOF
```

## License

Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
