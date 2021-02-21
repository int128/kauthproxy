# kauthproxy [![CircleCI](https://circleci.com/gh/int128/kauthproxy.svg?style=shield)](https://circleci.com/gh/int128/kauthproxy) ![e2e-test](https://github.com/int128/kauthproxy/workflows/e2e-test/badge.svg)

This is a kubectl plugin of the authentication proxy to access [Kubernetes Dashboard](https://github.com/kubernetes/dashboard).

![screenshot](https://github.com/int128/kauthproxy/wiki/refs/heads/master/screenshot.png)

This allows you to access Kubernetes Dashboard with authentication.
You no longer need to [enter a service account token in Kubernetes Dashboard](https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/creating-sample-user.md).
It provides better **user experience and security**.

kauthproxy supports the following environments:

- Amazon EKS
- Azure Kubernetes Service (with Azure AD)
- Self-hosted Kubernetes cluster
  - [OpenID Connect tokens authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#openid-connect-tokens)
  - [Webhook token authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication)
  - [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator)

Note that kauthproxy does not work with [client certificate authentication](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#x509-client-certs).


## Getting Started

### Install

Install the latest release from [Homebrew](https://brew.sh/), [Krew](https://github.com/kubernetes-sigs/krew) or [GitHub Releases](https://github.com/int128/kauthproxy/releases).

```sh
# Homebrew (macOS)
brew install int128/kauthproxy/kauthproxy

# Krew (macOS, Linux and Windows)
kubectl krew install auth-proxy
```

You can deploy the manifest of Kubernetes Dashboard from [here](https://github.com/kubernetes/dashboard).

### Run

To access Kubernetes Dashboard in your cluster:

```
% kubectl auth-proxy -n kubernetes-dashboard https://kubernetes-dashboard.svc
Starting an authentication proxy for pod/kubernetes-dashboard-57fc4fcb74-jjg77:8443
Open http://127.0.0.1:18000
Forwarding from 127.0.0.1:57866 -> 8443
Forwarding from [::1]:57866 -> 8443
```

It will automatically open the browser, and you will see Kubernetes Dashboard logged in as you.
You do not need to enter your token.


## How it works

### Authentication

Kubernetes Dashboard supports [header based authentication](https://github.com/kubernetes/dashboard/blob/master/docs/user/access-control/README.md#authorization-header).
kauthproxy forwards HTTP requests from the browser to Kubernetes Dashboard.

Take a look at the diagram:

![diagram](docs/kauthproxy.svg)

When you access Kubernetes Dashboard, kauthproxy forwards HTTP requests by the following process:

1. Acquire your token from the credential plugin or authentication provider.
1. Set `authorization: bearer TOKEN` header to a request and forward the request to the pod.

### Authorization

kauthproxy requires the following privileges:

- Get the Service of Kubernetes Dashboard.
- List the Pods of Kubernetes Dashboard.
- Port-forward to the Pod of Kubernetes Dashboard.

If you need to assign the least privilege for production,
see [an example of `Role`](e2e_test/kauthproxy-role.yaml).


## Usage

```
Usage:
  kubectl auth-proxy POD_OR_SERVICE_URL [flags]

Flags:
      --add_dir_header                   If true, adds the file directory to the header
      --address stringArray              The address on which to run the proxy. If set multiple times, it will try binding the address in order (default [127.0.0.1:18000,127.0.0.1:28000])
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
      --skip-open-browser                If set, skip opening the browser
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

### End-to-end test

To provision a cluster:

```sh
# requires Docker, Kind and Chrome
brew cask install docker google-chrome
brew install kind

# provision a cluster and deploy Kubernetes Dashboard
make -C e2e_test deploy
```

You can access the cluster as follows:

```sh
export KUBECONFIG=e2e_test/output/kubeconfig.yaml

# show all pods
kubectl get pods -A

# open Kubernetes Dashboard
./kauthproxy -n kubernetes-dashboard --user=tester https://kubernetes-dashboard.svc
```

To run the automated test:

```sh
make -C e2e_test test
```

To delete the cluster.

```sh
make -C e2e_test delete-cluster
```
