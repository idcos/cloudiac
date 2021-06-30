# Runner 配置项



runner:
  asset_path: "/tmp" // 该目录会在容器启动是挂载到容器的 `/asset` 目录，用于将terraform provider和ssh私钥映射到容器内使用
  log_base_path: "/tmp"  // 每个作业对应容器的日志会在该目录下创建，容器日志目录格式为：模板Guid/作业Guid，例如：tm-fffff/task-bbbbbbb；同时，需要把cmds/runner/state.tf.tmpl 模版文件提前放置于该配置的目录下
  default_image: "mt5225/tf-ansible:v0.0.1"  // 启动容器默认使用的镜像名称