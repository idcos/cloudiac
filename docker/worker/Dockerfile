FROM mini/ansible:2.4

ENV TERRAFORM_VERSION=0.14.11
RUN apk add --update git curl openssh && \
    curl https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip | \
        busybox unzip - && mv terraform /bin/ && chmod a+x /bin/terraform

COPY assets/terraform.py /cloudiac/assets/terraform.py
COPY assets/providers /usr/share/terraform/plugins

