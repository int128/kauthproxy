CLUSTER_NAME := kauthproxy-e2e
KUBECONFIG := output/kubeconfig.yaml
export KUBECONFIG

.PHONY: all
all:
	$(MAKE) deploy
	$(MAKE) test

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
	# Create a service account for end-to-end test
	kubectl apply -f kauthproxy-role.yaml
	kubectl get serviceaccount tester '-ojsonpath={.secrets[0].name}' | xargs kubectl get secret '-ojsonpath={.data.token}' | base64 --decode | xargs kubectl config set-credentials tester --token

.PHONY: test
test: main.go
	go run main.go

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster --name $(CLUSTER_NAME)
	-rm $(KUBECONFIG)

.PHONY: ci-publish
ci-publish:
	go install github.com/int128/ghcp
	ghcp commit -u int128 -r kauthproxy -b screenshot -m "ref: $(GITHUB_SHA)" -C output screenshot.png
