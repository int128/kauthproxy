# kubectl-oidc-proxy

This is a kubectl plugin of the reverse proxy with the `authorization` header.
It provides OpenID Connect authentication for Kubernetes Dashboard.

**Status:** Proof of concept. Not for production.


## Getting Started

```sh
go build -o kubectl-oidc_proxy
```

You need to setup Kubernetes OpenID Connect authentication and
install [kubelogin](https://github.com/int128/kubelogin).

```sh
export KUBECONFIG=.kubeconfig

# Login and update the kubeconfig
kubectl oidc-login

# Start a proxy to the Kubernetes Dashboard
kubectl -n kube-system port-forward svc/kubernetes-dashboard 8443:443

# Start a proxy to the above
kubectl oidc-proxy
```

Open http://localhost:8888 and then the Kubernetes Dashboard should appear.
