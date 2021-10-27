FROM centos:7

ARG WORKDIR=/usr/yunji/cloudiac
WORKDIR ${WORKDIR}

RUN yum install epel-release -y && \
    yum install git curl unzip openssh ansible -y && \
    yum clean all

ENV TERRASCAN_VERSION=1.9.0
RUN mkdir -p /root/.terrascan/pkg/policies/opa/rego && \
    curl -L https://github.com/accurics/terrascan/releases/download/v${TERRASCAN_VERSION}/terrascan_${TERRASCAN_VERSION}_Linux_x86_64.tar.gz | tar -xz terrascan && install terrascan /usr/local/bin && rm terrascan 

RUN git clone https://github.com/jinxing-idcos/tfenv.git /root/.tfenv && cd /root/.tfenv && git checkout tags/v2.2.3
ENV PATH="/root/.tfenv/bin:${PATH}"
RUN tfenv install "0.11.15" && \
    tfenv install "0.12.31" && \
    tfenv install "0.13.7" && \
    tfenv install "0.14.11" && \
    tfenv install "0.15.5" && \
    tfenv install "1.0.6"

COPY assets/terraform.py /cloudiac/assets/terraform.py
COPY assets/providers /usr/share/terraform/plugins

