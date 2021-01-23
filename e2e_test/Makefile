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
	kustomize build | kubectl apply -f -
	kubectl -n kubernetes-dashboard rollout status deployment kubernetes-dashboard
	kubectl -n kube-system rollout status deployment metrics-server
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
