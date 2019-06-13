# kubectl-oidc-port-forward

This is a kubectl plugin of the reverse proxy with the `authorization` header.
It provides OpenID Connect authentication for Kubernetes Dashboard.

**Status:** Proof of concept. Not for production.


## Getting Started

```sh
go build -o kubectl-oidc_port_forward
```

You need to setup Kubernetes OpenID Connect authentication and
install [kubelogin](https://github.com/int128/kubelogin).

```sh
export KUBECONFIG=.kubeconfig

# Login and update the kubeconfig
kubectl oidc-login

# Forward local port to the Kubernetes Dashboard service
kubectl oidc-port-forward svc/kubernetes-dashboard 8443:443
```

Open http://localhost:8888 and then the Kubernetes Dashboard should appear.
