# 容器部署

该文档使用容器部署了所有组件，包括 mysql 和 consul，实际生产环境请根据需要进行调整。

## 1. 安装并启动 docker

```bash
curl -fsSL https://get.docker.com | bash -s docker && \
systemctl enable docker && \
systemctl start docker
```

## 2. 安装 docker-compose

```bash
curl -L https://get.daocloud.io/docker/compose/releases/download/1.29.2/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose && \
chmod +x /usr/local/bin/docker-compose
```

## 3. 创建部署目录

!!! Caution
    **部署目录必须为 /usr/yunji/cloudiac，部署到其他目录将无法执行环境部署任务。**

```bash
mkdir -p /usr/yunji/cloudiac/var/{consul,mysql,acltls} && cd /usr/yunji/cloudiac/
```

## 4. 创建 docker-compose.yml 文件

文件路径 /usr/yunji/cloudiac/docker-compose.yml，内容如下:

```yaml
# auto-replace-from: docker/docker-compose.yml
version: "3.2"
services:
  iac-portal:
    container_name: iac-portal
    image: "${DOCKER_REGISTRY}cloudiac/iac-portal:v0.9.1"
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var
        target: /usr/yunji/cloudiac/var
      - type: bind
        source: /usr/yunji/cloudiac/.env
        target: /usr/yunji/cloudiac/.env
    ports:
      - "9030:9030"
    depends_on:
      - mysql
      - consul
    restart: always

  ct-runner:
    container_name: ct-runner
    image: "${DOCKER_REGISTRY}cloudiac/ct-runner:v0.9.1"
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var
        target: /usr/yunji/cloudiac/var
      - type: bind
        source: /usr/yunji/cloudiac/.env
        target: /usr/yunji/cloudiac/.env
      - type: bind
        source: /var/run/docker.sock
        target: /var/run/docker.sock
    ports:
      - "19030:19030"
    depends_on:
      - consul
    restart: always

  iac-web:
    container_name: iac-web
    image: "${DOCKER_REGISTRY}cloudiac/iac-web:v0.9.1"
    ports:
      - 80:80
    restart: always
    depends_on:
      - iac-portal

  mysql:
    container_name: mysql
    image: "mysql:8.0"
    command: [
        "--character-set-server=utf8mb4",
        "--collation-server=utf8mb4_unicode_ci",
        "--sql_mode=STRICT_TRANS_TABLES,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"
    ]
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var/mysql
        target: /var/lib/mysql
    environment:
      - MYSQL_RANDOM_ROOT_PASSWORD=yes
      - MYSQL_USER
      - MYSQL_PASSWORD
      - MYSQL_DATABASE
    restart: always

  consul:
    container_name: consul
    image: "consul:latest"
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var/consul
        target: /consul/data   
    ports:
      - "8500:8500"
    command: >
      consul agent -server -bootstrap-expect=1 -ui -bind=0.0.0.0
      -client=0.0.0.0 -enable-script-checks=true -data-dir=/consul/data 
    restart: always

```

> 开启acl和tls部署的docker-compose.yaml

文件路径 /usr/yunji/cloudiac/docker-compose.yml，内容如下:
```yaml
# auto-replace-from: docker/docker-compose.yml
version: "3.2"
services:
  iac-portal:
    container_name: iac-portal
    image: "${DOCKER_REGISTRY}cloudiac/iac-portal:v0.9.1"
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var
        target: /usr/yunji/cloudiac/var
      - type: bind
        source: /usr/yunji/cloudiac/.env
        target: /usr/yunji/cloudiac/.env
    ports:
      - "9030:9030"
    depends_on:
      - mysql
      - consul
    restart: always

  ct-runner:
    container_name: ct-runner
    image: "${DOCKER_REGISTRY}cloudiac/ct-runner:v0.9.1"
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var
        target: /usr/yunji/cloudiac/var
      - type: bind
        source: /usr/yunji/cloudiac/.env
        target: /usr/yunji/cloudiac/.env
      - type: bind
        source: /var/run/docker.sock
        target: /var/run/docker.sock
    ports:
      - "19030:19030"
    depends_on:
      - consul
    restart: always

  iac-web:
    container_name: iac-web
    image: "${DOCKER_REGISTRY}cloudiac/iac-web:v0.9.1"
    ports:
      - 80:80
    restart: always
    depends_on:
      - iac-portal

  mysql:
    container_name: mysql
    image: "mysql:8.0"
    command: [
        "--character-set-server=utf8mb4",
        "--collation-server=utf8mb4_unicode_ci",
        "--sql_mode=STRICT_TRANS_TABLES,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"
    ]
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var/mysql
        target: /var/lib/mysql
    environment:
      - MYSQL_RANDOM_ROOT_PASSWORD=yes
      - MYSQL_USER
      - MYSQL_PASSWORD
      - MYSQL_DATABASE
    restart: always

  consul:
    container_name: consul
    image: "consul:latest"
    environment:
      - CONSUL_HTTP_SSL_VERIFY=false
      - CONSUL_HTTP_SSL=true
    volumes:
      - type: bind
        source: /usr/yunji/cloudiac/var/consul
        target: /consul/data
      - type: bind
        source: /usr/yunji/cloudiac/var/acltls
        target: /consul/config    
    ports:
      - "8500:8500"
    command: >
      consul agent -server -bootstrap-expect=1 -ui -bind=0.0.0.0
      -client=0.0.0.0 -enable-script-checks=true -data-dir=/consul/data -config-dir=/consul/config
    restart: always
```

开启consul配置acl
```bash
# 新增consul配置acl.hcl
cat >> /usr/yunji/cloudiac/var/acltls/acl.hcl <<EOF
acl = {
  enabled = true
  default_policy = "deny"
  enable_token_persistence = true
}
EOF

# 重启consul
docker restart consul

# 进入容器
docker exec -it consul sh

# 生成token,保存好生成的SecretID
consul acl bootstrap

#退出容器

# 加入SecretID作为token加入acl.hcl配置
cat > /usr/yunji/cloudiac/var/acltls/acl.hcl <<EOF
acl = {
  enabled = true
  default_policy = "deny"
  enable_token_persistence = true
  tokens {
    master = "5ecc86f0-fa68-6ddc-e848-ef382d7737ec" #SecretID
  }
}
EOF
```


开启consul配置tls访问
> 证书名称固定 ca.pem,client.key,client.pem

```bash
cd /usr/yunji/cloudiac/var/acltls

 #生成根证书key
openssl genrsa -out ca.key 2048
#生成根证书密钥
openssl req -new -x509 -days 7200 -key ca.key   -out ca.pem

#生成客户端私钥
openssl genrsa -out client.key 2048

#生成的客户端的CSR
openssl req -new -key client.key  -out client.csr

# 创建宿舍机对应的签名证书  IP宿主机
echo subjectAltName = IP:127.0.0.1 > extfile.cnf

#客户端自签名的证书
openssl x509 -req -days 365 -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
   -out client.pem -extfile extfile.cnf
   
# 新增consul配置tls.json
cat >> /usr/yunji/cloudiac/var/acltls <<EOF
{
  "verify_incoming": false,
  "verify_incoming_rpc": true,
  "ports": {
    "http": -1,
    "https": 8500
  },
  "ca_file": "/consul/config/ca.pem",
  "cert_file": "/consul/config/client.pem",
  "key_file": "/consul/config/client.key"
}
EOF
```

## 5. 创建 .env 文件

文件路径 /usr/yunji/cloudiac/.env，内容如下:

```bash
# auto-replace-from: configs/dotenv.sample
# 平台管理员账号密码(均为必填)
# 该账号密码只在系统初始化时使用，后续修改不影响己创建的账号
IAC_ADMIN_EMAIL="admin@example.com"
# 密码要求长度大于 8 且包含字母、数字、特殊字符
IAC_ADMIN_PASSWORD=""

# 加密密钥配置(必填)
# 敏感数据使用该密钥进行加密
SECRET_KEY=""

# IaC 对外提供服务的地址(必填), 示例: http://cloudiac.example.com
# 该地址需要带协议(http/https)，结尾不可以加 "/"
PORTAL_ADDRESS=""

# consul 地址(必填)，示例: private.host.ip:8500
# 需要配置为机器的内网 ip:port，不可使用 127.0.0.1
CONSUL_ADDRESS=""

# cloudiac registry 服务地址(选填)，示例：http://registry.cloudiac.org
REGISTRY_ADDRESS=""

# 使用 https 向外（比如runner）发送请求的时候是否允许使用不安全证书
HTTP_CLIENT_INSECURE=false

# mysql 配置(必填)
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_DATABASE=cloudiac
MYSQL_USER=cloudiac
MYSQL_PASSWORD="mysqlpass"

# portal 服务注册信息配置 (均为必填)
## portal 服务的 IP 地址， 容器化部署时无需修改, 手动部署时配置为内网 IP
SERVICE_IP=iac-portal
## portal 服务注册的 id(需要保证唯一)
SERVICE_ID=iac-portal-01
## portal 服务注册的 tags
SERVICE_TAGS="iac-portal;portal-01"

# docker reigstry 地址，为空时使用 docker hub
DOCKER_REGISTRY=""

# logger 配置
LOG_DEVEL="info"

# SMTP 配置(该配置只影响邮件通知的发送，不配置不影响其他功能)
## example: smtp.example.com:25
SMTP_ADDRESS=""
## example: user@example.com
SMTP_USERNAME=""
SMTP_PASSWORD=""
## example: support@example.com
SMTP_FROM=""
SMTP_FROM_NAME=IaC

# KAFKA配置，配置后每次执行部署任务都会将环境的最新全量资源详情通过 kafka 消息发送
KAFKA_TOPIC="IAC_TASK_REPLY"
KAFKA_GROUP_ID=""
KAFKA_PARTITION=0
## example: KAFKA_BROKERS: ["kafka.example.com:9092", "..."]
KAFKA_BROKERS=[]
KAFKA_SASL_USERNAME=""
KAFKA_SASL_PASSWORD=""

######### 以下为 runner 配置 #############
# runner 服务注册配置(均为必填)
## runner 服务的 IP 地址， 容器化部署时无需修改, 手动部署时配置为内网 IP
RUNNER_SERVICE_IP=ct-runner
## runner 服务注册的 id(需要保证唯一)
RUNNER_SERVICE_ID=ct-runner-01
RUNNER_SERVICE_TAGS="ct-runner;runner-01"

## 是否开启 offline mode，默认为 false
RUNNER_OFFLINE_MODE="false"
```

!!! Caution
    `.env` 中以下配置为**必填项**，其他配置可根据需要修改：

    - IAC_ADMIN_PASSWORD: 初始的平台管理员密码
    - SECRET_KEY: 数据加密存储时使用的密钥
    - PORTAL_ADDRESS: 对外地址服务的地址
    - CONSUL_ADDRESS: consul 服务地址，配置为部署机内网 ip:8500 端口即可

!!! Info
    通过 `.env` 可以实现大部分配置的修改，更多配置项可查看 docker 镜像中的 config-portal.yml 和 config-runner.yml 文件，需要修改可以从镜像中拷贝文件，修改后再在容器启动时挂载进行替换。

## 6. 启动docker-compose

```bash
docker-compose up
```

> 默认为前台启动，以便于排查问题，在确定服务正常后可以改为后台启动：`dokcker-compose up -d`。


## 7. 部署完成
至此服务部署完成，访问 http://${PORTAL_ADDRESS} 进行登陆。

默认的用户名为 admin@example.com (即 IAC_ADMIN_EMAIL)，密码为 `.env` 中配置的 `IAC_ADMIN_PASSWORD`。
