# kauthproxy [![CircleCI](https://circleci.com/gh/int128/kauthproxy.svg?style=shield)](https://circleci.com/gh/int128/kauthproxy)

This is a kubectl plugin to forward a local port to a pod or service via authentication proxy.
It gets a token from the credential plugin (e.g. [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) or [kubelogin](https://github.com/int128/kubelogin)) and forwards requests to a pod or service with `Authorization: Bearer token` header.

Take a look at the concept:

```
+--------------------------------+
| Browser                        |
+--------------------------------+
  ↓ http://localhost:random_port
+--------------------------------+              +-----------------------------+
| kubectl auth-proxy             | <-- token -- | client-go credential plugin |
+--------------------------------+              +-----------------------------+
  ↓ https://localhost:443
  ↓ Authorization: Bearer token
+--------------------------------+
| Kubernetes Dashboard (service) |
+--------------------------------+
```

**Status**: alpha and not for production.


## Getting Started

### Install

You can install the latest release from [Homebrew](https://brew.sh/), [Krew](https://github.com/kubernetes-sigs/krew) or [GitHub Releases](https://github.com/int128/kauthproxy/releases) as follows:

```sh
# Homebrew
brew tap int128/kauthproxy
brew install kauthproxy

# Krew (TODO)
kubectl krew install auth-proxy

# GitHub Releases
curl -LO https://github.com/int128/kauthproxy/releases/download/v0.1.0/kauthproxy_linux_amd64.zip
unzip kauthproxy_linux_amd64.zip
ln -s kauthproxy kubectl-auth_proxy
```


### Kubernetes Dashboard on Amazon EKS

You need to [configure the kubeconfig](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html) to use [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) or `aws eks get-token`.

To run an authentication proxy to the service:

```
% kubectl auth-proxy -n kube-system https://kubernetes-dashboard.svc
Open http://127.0.0.1:57777
Forwarding from 127.0.0.1:57776 -> 8443
Forwarding from [::1]:57776 -> 8443
```

Open the URL and you can access the Kubernetes Dashboard with the token.


### Kubernetes Dashboard with OpenID Connect authentication

You need to configure the kubeconfig to use [kubelogin](https://github.com/int128/kubelogin).

Run the following command,

```
% kubectl auth-proxy -n kube-system https://kubernetes-dashboard.svc
Open http://127.0.0.1:57777
Forwarding from 127.0.0.1:57776 -> 8443
Forwarding from [::1]:57776 -> 8443
```

Open the URL and you can access the Kubernetes Dashboard with the token.


### Kibana with OpenID Connect authentication

You need to configure the kubeconfig to use [kubelogin](https://github.com/int128/kubelogin).

Run the following command,

```
% kubectl auth-proxy https://kibana
Open http://127.0.0.1:57777
Forwarding from 127.0.0.1:57776 -> 8443
Forwarding from [::1]:57776 -> 8443
```

Open the URL and you can access the Kibana with the token.


## Known Issues

- kauthproxy always skips TLS verification for a pod. TODO: add a flag


## Usage

```
Forward a local port to a pod or service via authentication proxy.
To forward a local port to a service, set a service name with .svc suffix. e.g. http://service-name.svc
To forward a local port to a pod, set a pod name. e.g. http://pod-name

Usage:
  kubectl auth-proxy POD_OR_SERVICE_URL [flags]

Examples:
  kubectl auth-proxy https://kubernetes-dashboard.svc

Flags:
      --address string                   The address on which to run the proxy. Default to a random port of localhost. (default "localhost:0")
      --alsologtostderr                  log to standard error as well as files
      --as string                        Username to impersonate for the operation
      --as-group stringArray             Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --cache-dir string                 Default HTTP cache directory (default "~/.kube/http-cache")
      --certificate-authority string     Path to a cert file for the certificate authority
      --client-certificate string        Path to a client certificate file for TLS
      --client-key string                Path to a client key file for TLS
      --cluster string                   The name of the kubeconfig cluster to use
      --context string                   The name of the kubeconfig context to use
  -h, --help                             help for kubectl
      --insecure-skip-tls-verify         If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string                Path to the kubeconfig file to use for CLI requests.
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --log_file string                  If non-empty, use this log file
      --log_file_max_size uint           Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files (default true)
  -n, --namespace string                 If present, the namespace scope for this CLI request
      --request-timeout string           The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                    The address and port of the Kubernetes API server
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
      --token string                     Bearer token for authentication to the API server
      --user string                      The name of the kubeconfig user to use
  -v, --v Level                          number for the log level verbosity
      --version                          version for kubectl
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```


## Contributions

This is an open source software.
Feel free to open issues and pull requests.
