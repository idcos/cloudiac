FROM mini/ansible:2.4

ENV TERRAFORM_VERSION=1.0.0
RUN apk add --update git curl openssh && \
    curl https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip | \
        busybox unzip - && mv terraform /bin/ && chmod a+x /bin/terraform

COPY assets/terraform.py /cloud-iac/assets/terraform.py
COPY assets/providers /usr/share/terraform/plugins
