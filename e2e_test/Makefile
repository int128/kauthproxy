CLUSTER_NAME := kauthproxy-e2e
KUBECONFIG := output/kubeconfig.yaml
export KUBECONFIG

.PHONY: all
all: test

.PHONY: cluster
cluster: $(KUBECONFIG)
$(KUBECONFIG):
	kind create cluster --name $(CLUSTER_NAME)

.PHONY: deploy
deploy: cluster
	kustomize build | kubectl apply -f -
	kubectl -n kube-system rollout status deployment metrics-server
	kubectl -n kubernetes-dashboard rollout status deployment kubernetes-dashboard
	kubectl get secret tester-token '-ojsonpath={.data.token}' | base64 --decode | xargs kubectl config set-credentials tester --token

output/kauthproxy:
	go build -o $@ ..

output/e2e_test:
	go build -o $@ .

.PHONY: test
test: deploy output/e2e_test output/kauthproxy
	./output/e2e_test

.PHONY: delete-cluster
delete-cluster:
	kind delete cluster --name $(CLUSTER_NAME)
	-rm $(KUBECONFIG)

.PHONY: ci-publish-screenshot
ci-publish-screenshot:
	./publish-screenshot.sh
