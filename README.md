# kubectl-oidc-port-forward [![CircleCI](https://circleci.com/gh/int128/kubectl-oidc-port-forward.svg?style=shield)](https://circleci.com/gh/int128/kubectl-oidc-port-forward)

This is a kubectl plugin for [Kubernetes OpenID Connect (OIDC) authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens), which provides port forwarding with the `authorization` header.

You can access to the [Kubernetes Dashboard](https://github.com/kubernetes/dashboard) using OIDC authentication via a tunnel established by this plugin.

```
+---------------------------+
| Browser                   |
+---------------------------+
  ↓ http://localhost:8888
+---------------------------+
| kubectl oidc-port-forward | This adds the authorization header.
+---------------------------+
  ↓ https://localhost:x
+---------------------------+
| kubectl port-forward      | This forwards requests to the service.
+---------------------------+
  ↓ TCP
+---------------------------+
| svc/kubernetes-dashboard  |
+---------------------------+
```

**Status:** Proof of concept. Not for production.


## Getting Started

You need to configure the OIDC provider, Kubernetes API server, kubectl authentication and role binding.

```sh
# Point the kubeconfig
export KUBECONFIG=.kubeconfig

# Login to the OIDC provider
kubectl oidc-login

# Forward the local port to the Kubernetes Dashboard service
kubectl oidc-port-forward svc/kubernetes-dashboard 8888:https/443
```

Open http://localhost:8888 and then the Kubernetes Dashboard should be shown.
