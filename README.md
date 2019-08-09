# kubectl-auth-port-forward [![CircleCI](https://circleci.com/gh/int128/kubectl-auth-port-forward.svg?style=shield)](https://circleci.com/gh/int128/kubectl-auth-port-forward)

This is a kubectl plugin for port forwarding with an `authorization` header.

You can access to the [Kubernetes Dashboard](https://github.com/kubernetes/dashboard) using OIDC authentication via a tunnel established by this plugin.

```
+---------------------------+
| browser                   |
+---------------------------+
  ↓ http://localhost:8000
+---------------------------+     +-----------------------------+
| kubectl auth-port-forward | <-> | client-go credential plugin |
+---------------------------+     +-----------------------------+
  ↓ https://localhost:443
+---------------------------+
| svc/kubernetes-dashboard  |
+---------------------------+
```

**Status:** Proof of concept. Not for production.


## Getting Started

You need to set the credential plugin in the kubeconfig.

If you are using [aws eks get-token](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html),

```yaml
users:
- name: iam
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - CLUSTER_NAME
```

If you are using [kubelogin](https://github.com/int128/kubelogin),

```yaml
users:
- name: keycloak
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl
      args:
      - oidc-login
      - get-token
      - --oidc-issuer-url=https://issuer.example.com
      - --oidc-client-id=YOUR_CLIENT_ID
      - --oidc-client-secret=YOUR_CLIENT_SECRET
```

Run the plugin.

```sh
kubectl auth-port-forward -n kubernetes-dashboard kubernetes-dashboard-xxx 8080:https/8443
```

Open http://localhost:8080 and then the Kubernetes Dashboard should be shown.
