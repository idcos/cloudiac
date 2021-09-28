# 创建演示组织
该文档用于快速创建一个演示组织及其基础配置，创建完成后可以在演示组织中快速创建环境进行体验。

以下操作均在 cloudiac 部署目录 `/usr/yunji/cloudiac` 下执行。

#### 1. 准备 SSH 密钥对
创建演示组织需要为其添加 SSH 密钥己便执行 playbook。

您可以使用己创建的 SSH 密钥或者生成一个新的密钥，并将私钥保存为 "var/private_key"

#### 2. 创建演示组织配置文件:
 创建文件 `demo-conf.yml`，该文件用于定义演示组织需要进行的配置，具体内容及配置说明如下:

```yaml
organization:
  name: 演示组织
  description: 这是一个演示组织
  project:
    name: 演示项目
    description: 这是一个演示项目

  # SSH 私钥文件路径(该私钥会被添加为云商 SSH 密钥对，并绑定到创建的计算资源)
  key:
    name: 演示密钥
    key_file: ./var/private_key

  # 为组织添加变量
  variables:
    # 云商认证密钥可以通过环境变量传入，以下示例为添加 aliyun 的 API 密钥
    - name: ALICLOUD_ACCESS_KEY
      type: environment   # 变量类型为 environment(即环境变量)
      value: xxxxxxxxxx 
      description: 阿里云访问密钥ID
      sensitive: false    
    - name: ALICLOUD_SECRET_KEY
      type: environment
      value: xxxxxxxxxx
      description: 阿里云访问密钥
      sensitive: true # 是否为敏感变量(敏感变量会加密保存，且前端不展示)

  template:
    name: 演示云模板
    description: 这是一个内置的示例模板，该模板在 aliyun 上创建一台 ec2，
        绑定公网 IP 并安装 nginx 服务，
        部署后可直接通过 outputs 中的 ip 访问
    repo_id: /cloudiac/cloudiac-example.git 
    revision: master
    tf_vars_file: qa-env.tfvars
    playbook: ansible/playbook.yml
    workdir:
```

#### 3. 将文件挂载到容器
> **该步骤只在使用容器化部署时需要执行**

修改 docker-compose.yml, 为 service iac-portal 追加 volumes:
```yaml
#services:
#  iac-portal:
#    container_name: iac-portal
#    image: "cloudiac/iac-portal:latest"
#    volumes:
#      - type: bind
#        source: /usr/yunji/cloudiac/var
#        target: /usr/yunji/cloudiac/var
#      - type: bind
#        source: /usr/yunji/cloudiac/.env
#        target: /usr/yunji/cloudiac/.env

      # 以下为追加内容
      - type: bind
        source: /usr/yunji/cloudiac/var/private_key
        target: /usr/yunji/cloudiac/var/private_key
      - type: bind
        source: /usr/yunji/cloudiac/demo-conf.yml
        target: /usr/yunji/cloudiac/demo-conf.yml
# ...
```

修改完成后执行以下命令，重启 iac-portal 使配置生效:

```bash
docker-compose up -d iac-portal --force-recreate
```

#### 4. 生成演示组织
使用 `iac-tool` 自动创建演示组织。

如果您是**容器化部署**的环境，则执行以下命令:
```bash
docker-compose exec iac-portal ./iac-tool init-demo demo-conf.yml
```

如果您是**手动部署**的环境，则执行以下命令:
```bash
./iac-tool init-demo demo-conf.yml
```

**⚠️注意，演示组织只能创建一次，如需重新初始化，需手动删除相关数据**

#### 5. 重启 iac-portal 服务

**容器化部署**的环境执行以下命令:
```bash
docker-compose restart iac-portal
```

**手动部署**的环境执行以下命令:
```bash
systemctl restart iac-portal
```

*至此演示组织创建完成，登录系统创建环境进行体验吧。*