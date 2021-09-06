FROM mini/ansible:2.4

ARG WORKDIR=/root
WORKDIR ${WORKDIR}
ENV TERRASCAN_VERSION=1.9.0
ENV PATH="/root/.tfenv/bin:${PATH}"
RUN apk add --update git bash curl && \
    git clone https://github.com/tfutils/tfenv.git /root/.tfenv && \
    tfenv install "v0.11.15" && \
    tfenv install "v0.12.31" && \
    tfenv install "v0.13.7" && \
    tfenv install "v0.14.11" && \
    tfenv install "v0.15.5" && \
    tfenv install "v1.0.5" && \
    curl -L https://github.com/accurics/terrascan/releases/download/v${TERRASCAN_VERSION}/terrascan_${TERRASCAN_VERSION}_Linux_x86_64.tar.gz | \
        tar -xz terrascan && install terrascan /usr/local/bin && rm terrascan

COPY assets/terraform.py /cloudiac/assets/terraform.py
COPY assets/providers /usr/share/terraform/plugins
