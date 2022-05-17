FROM cloudiac/base-ct-runner:v0.1.5

ARG WORKER_IMAGE=cloudiac/ct-worker:latest
EXPOSE 19030

COPY configs/config-runner.yml.sample config-runner.yml
COPY build/ct-runner ct-runner

# runner 容器化部署时不能设置 assets_path
# 同时如果指定了 WORKER_IMAGE，则做配置替换
RUN sed -i -e 's/ assets_path:.*/ assets_path: ""/' \
           -e 's#cloudiac/ct-worker:latest"#'$WORKER_IMAGE'"#g' config-runner.yml

CMD ./ct-runner
