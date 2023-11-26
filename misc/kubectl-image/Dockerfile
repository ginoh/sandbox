FROM debian:bullseye-slim

ARG KUBECTL_VERSION

RUN apt-get update && apt-get install -y \
    curl \
    gettext \
    git \
    jq \
    #openssh-server \
    python3-dev \
    python3-pip && \    
    rm -rf /var/lib/apt/lists/* && \
    if [ "${KUBECTL_VERSION}" = "" ]; then curl -L -s -o /tmp/kubectl_version https://dl.k8s.io/release/stable.txt; else echo -n "${KUBECTL_VERSION}" > /tmp/kubectl_version; fi && \
    curl -L -o kubectl "https://dl.k8s.io/release/$(cat /tmp/kubectl_version)/bin/linux/amd64/kubectl" && \
    curl -L -o kubectl-argo-rollouts https://github.com/argoproj/argo-rollouts/releases/latest/download/kubectl-argo-rollouts-linux-amd64 && \
    curl -L -o yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64 && \
    chmod +x kubectl kubectl-argo-rollouts yq && \
    mv kubectl kubectl-argo-rollouts yq /usr/local/bin && \
    rm /tmp/kubectl_version && \
    kubectl version --client && \
    kubectl argo rollouts version

VOLUME ["/root/.kube"]