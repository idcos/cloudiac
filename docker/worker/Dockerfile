FROM cloudiac/base-ct-worker:v0.1.7

RUN yum install python-requests -y

COPY assets/terraform.py /cloudiac/assets/terraform.py
COPY assets/terraformrc-* /cloudiac/assets/
COPY assets/cloudiac-playbook /usr/local/bin/cloudiac-playbook
COPY assets/saltapi.py /usr/share/ansible/plugins/connection/saltapi.py
COPY build/iac-tool /usr/yunji/cloudiac/iac-tool

