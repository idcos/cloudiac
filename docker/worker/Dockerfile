FROM mini/ansible:2.4

ARG WORKDIR=/usr/yunji/cloudiac
WORKDIR ${WORKDIR}
RUN apk add --update git bash curl && \
    git clone https://github.com/tfutils/tfenv.git /root/.tfenv

ENV PATH="/root/.tfenv/bin:${PATH}"
RUN tfenv install "0.11.15" && \
    tfenv install "0.12.31" && \
    tfenv install "0.13.7" && \
    tfenv install "0.14.11" && \
    tfenv install "0.15.5" && \
    tfenv install "1.0.6"

COPY assets/terraform.py /cloudiac/assets/terraform.py
COPY assets/providers /usr/share/terraform/plugins
