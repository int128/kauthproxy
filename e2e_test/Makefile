CLUSTER_NAME := kauthproxy-e2e
KUBECONFIG := output/kubeconfig.yaml
export KUBECONFIG

.PHONY: all
all: deploy

.PHONY: cluster
cluster: $(KUBECONFIG)
$(KUBECONFIG):
	kind create cluster --name $(CLUSTER_NAME)

.PHONY: deploy
deploy: cluster
	# Deploy https://github.com/kubernetes/dashboard
	kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.3/aio/deploy/recommended.yaml
	kubectl -n kubernetes-dashboard rollout status deployment kubernetes-dashboard
	# Deploy https://github.com/kubernetes-sigs/metrics-server
	kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.3.7/components.yaml
	kubectl -n kube-system patch --type=json -p '[{"op": "add", "path": "/spec/template/spec/containers/0/args/0", "value": "--kubelet-insecure-tls"}]' deployment metrics-server
	kubectl -n kube-system rollout status deployment metrics-server
	# Create a service account for token authentication,
	# because kind provides cluster-admin with certificate authentication but kauthproxy requires token authentication
	kubectl create serviceaccount tester
	kubectl create clusterrolebinding tester --clusterrole=cluster-admin --serviceaccount=default:tester
	kubectl config set-credentials tester --token=$$(kubectl get secret $$(kubectl get serviceaccount tester '-ojsonpath={.secrets[0].name}') '-ojsonpath={.data.token}' | base64 --decode)

.PHONY: test
test: main.go
	go run main.go

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster --name $(CLUSTER_NAME)
	-rm $(KUBECONFIG)
