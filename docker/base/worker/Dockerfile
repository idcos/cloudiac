ARG PROVIDERS
FROM ${PROVIDERS} as providers

FROM centos:7

ARG WORKDIR=/usr/yunji/cloudiac
WORKDIR ${WORKDIR}
USER root

RUN yum install epel-release -y && \
    yum install -y centos-release-scl && \
    yum install git curl unzip openssh vim which -y && \
    yum install sshpass -y && \
    yum install openssl-1.0.2k -y && \
    yum install -y rh-python38 && \
    ln -s /opt/rh/rh-python38/root/usr/bin/python3 /usr/bin/python3 && \
    ln -s /opt/rh/rh-python38/root/usr/bin/pip3 /usr/bin/pip3 && \
    pip3 install ansible && \
    pip3 install urllib3 == 1.26.13 && \
    pip3 install pywinrm && \
    ln -s /opt/rh/rh-python38/root/usr/local/bin/ansible /usr/bin/ansible && \
    ln -s /opt/rh/rh-python38/root/usr/local/bin/ansible-playbook /usr/bin/ansible-playbook && \
    ln -s /opt/rh/rh-python38/root/usr/local/bin/ansible-galaxy /usr/bin/ansible-galaxy && \
    ln -s /opt/rh/rh-python38/root/usr/local/bin/ansible-inventory /usr/bin/ansible-inventory && \
    ln -s /opt/rh/rh-python38/root/usr/local/bin/ansible-pull /usr/bin/ansible-pull && \
    sed -i 's/#\s*StrictHostKeyChecking ask/    StrictHostKeyChecking no/' /etc/ssh/ssh_config && \
    yum clean all

RUN

ENV TERRASCAN_VERSION=1.9.0
RUN mkdir -p /root/.terrascan/pkg/policies/opa/rego && \
    curl -L https://github.com/accurics/terrascan/releases/download/v${TERRASCAN_VERSION}/terrascan_${TERRASCAN_VERSION}_Linux_x86_64.tar.gz | tar -xz terrascan && install terrascan /usr/local/bin && rm terrascan 

ENV TFSUMMARIZE_VERSION=0.3.2
RUN curl -L https://github.com/dineshba/tf-summarize/releases/download/v${TFSUMMARIZE_VERSION}/tf-summarize_linux_amd64.zip -O && unzip -o tf-summarize_linux_amd64.zip -d tf-summarize && install tf-summarize/tf-summarize /usr/local/bin && rm -rf tf-summarize

RUN git clone https://github.com/jinxing-idcos/tfenv.git /root/.tfenv && cd /root/.tfenv && git checkout tags/v2.2.3
ENV PATH="/root/.tfenv/bin:${PATH}"
RUN tfenv install "0.11.15" && \
    tfenv install "0.12.31" && \
    tfenv install "0.13.7" && \
    tfenv install "0.14.11" && \
    tfenv install "0.15.5" && \
    tfenv install "1.0.6" && \
    tfenv install "1.1.9" && \
    tfenv install "1.2.4" && \
    tfenv install "1.5.2" && \
    tfenv install "1.5.4" && \
    tfenv install "1.5.5" && \
    tfenv install "1.5.6"

RUN tfenv use 1.5.6 && \
  ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=cloudiac/base-ct-worker:v0.1.8 /cloudiac/terraform/plugins /cloudiac/terraform/plugins

