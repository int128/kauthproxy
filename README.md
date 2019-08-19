# kauthproxy [![CircleCI](https://circleci.com/gh/int128/kauthproxy.svg?style=shield)](https://circleci.com/gh/int128/kauthproxy)

This is a kubectl plugin of authentication proxy to a pod on Kubernetes.
It consists from the reverse proxy and port forwarder.

**Status**: alpha and not for production.

Take a look at the concept:

```
+--------------------------------+
| Browser                        |
+--------------------------------+
  ↓ http://localhost:8000
+--------------------------------+              +-----------------------------+
| kubectl auth-proxy             | <-- TOKEN -- | client-go credential plugin |
+--------------------------------+              +-----------------------------+
  ↓ https://localhost:443
  ↓ Authorization: Bearer TOKEN
+--------------------------------+
| Kubernetes Dashboard (service) |
+--------------------------------+
```


## Getting Started

### Kubernetes Dashboard on Amazon EKS

You need to [configure the kubeconfig](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html) to use aws-iam-authenticator or `aws eks get-token`.

To run an authentication proxy to the service:

```sh
kubectl auth-proxy -n kube-system kubernetes-dashboard-xxx 8000:https/8443
```

Open http://localhost:8000 and you can access the Kubernetes Dashboard with the token.


### Kubernetes Dashboard with OpenID Connect authentication

You need to configure the kubeconfig to use [`kubectl oidc-login`](https://github.com/int128/kubelogin).

Run the following command,

```sh
kubectl auth-proxy -n kube-system kubernetes-dashboard-xxx 8000:https/8443
```

Open http://localhost:8000 and you can access the Kubernetes Dashboard with the token.


### Kibana with OpenID Connect authentication

You need to configure the kubeconfig to use [`kubectl oidc-login`](https://github.com/int128/kubelogin).

Run the following command,

```sh
kubectl auth-proxy kibana-xxx 8000:http/4180
```

Open http://localhost:8000 and you can access the Kibana with the token.


## Usage

```
Forward a local port to a pod

Usage:
  kubectl auth-proxy POD_NAME LOCAL_PORT:POD_SCHEME/POD_PORT [flags]

Examples:
  kubectl -n kube-system auth-proxy kubernetes-dashboard-xxx 8443:https/8443

Flags:
      --as string                      Username to impersonate for the operation
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string               Default HTTP cache directory (default "~/.kube/http-cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
  -h, --help                           help for kubectl
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
  -n, --namespace string               If present, the namespace scope for this CLI request
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --version                        version for kubectl
```
