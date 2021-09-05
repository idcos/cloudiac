FROM mini/ansible:2.4

ARG WORKDIR=/root
WORKDIR ${WORKDIR}
RUN apk add --update git bash curl && \
    git clone https://github.com/tfutils/tfenv.git

ENV PATH="/root/tfenv/bin:${PATH}"
RUN tfenv install "v0.11.15" && \
    tfenv install "v0.12.31" && \
    tfenv install "v0.13.7" && \
    tfenv install "v0.14.11" && \
    tfenv install "v0.15.5" && \
    tfenv install "v1.0.5"

COPY assets/terraform.py /cloudiac/assets/terraform.py
COPY assets/providers /usr/share/terraform/plugins
