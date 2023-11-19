FROM docker:dind

ENV PATH /usr/local/go/bin:${PATH}
ENV GOLANG_VERSION 1.21.4

RUN set -eux; \
	arch="$(apk --print-arch)"; \
	case "$arch" in \
		'x86_64') \
			url='https://dl.google.com/go/go1.21.4.linux-amd64.tar.gz'; \
			sha256='73cac0215254d0c7d1241fa40837851f3b9a8a742d0b54714cbdfb3feaf8f0af'; \
			;; \
		'aarch64') \
			url='https://dl.google.com/go/go1.21.4.linux-arm64.tar.gz'; \
			sha256='ce1983a7289856c3a918e1fd26d41e072cc39f928adfb11ba1896440849b95da'; \
			;; \
		*) echo >&2 "error: unsupported architecture '${arch}' (likely packaging update needed)"; exit 1 ;; \
	esac; \
    \
	wget -O go.tgz "${url}"; \
	echo "${sha256} *go.tgz" | sha256sum -c -; \
    tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	go version;

ENV GOTOOLCHAIN=local
ENV GOPATH /go
ENV PATH ${GOPATH}/bin:${PATH}
RUN mkdir -p "${GOPATH}/src" "${GOPATH}/bin" && chmod -R 1777 "${GOPATH}"
WORKDIR ${GOPATH}
