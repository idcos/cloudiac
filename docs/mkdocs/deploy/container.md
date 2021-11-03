# 容器部署
该文档使用容器部署了所有组件，包括 mysql 和 consul，实际生产环境请根据需要进行调整。

#### 1. 安装并启动 docker
```bash
curl -fsSL https://get.docker.com | bash -s docker
systemctl enable docker
systemctl start docker
```

#### 2. 安装 docker-compose
```bash
curl -L https://get.daocloud.io/docker/compose/releases/download/1.29.2/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose && \
chmod +x /usr/local/bin/docker-compose
```

#### 3. 创建部署目录
部署目录固定为 /usr/yunji/cloudiac，**不可更改**，更改部署目录会导致执行部署任务失败。
```bash
mkdir -p /usr/yunji/cloudiac/var/{consul,mysql} && cd /usr/yunji/cloudiac/
```

#### 4. 编写 docker-compose.yml 文件
文件路径 /usr/yunji/cloudiac/docker-compose.yml，内容如下:
```yaml
version: "3.2"
services:
  iac-portal:
    container_name: iac-portal
    image: "cloudiac/iac-portal:latest"
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
    image: "cloudiac/ct-runner:latest"
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

  ct-worker:
    image: "cloudiac/ct-worker:latest"
    container_name: "ct-worker-say-hi"
    # 添加该服务只是为了自动 pull 镜像，并不需要后台运行
    command: ["echo", "hello world!"]
    restart: "no"

  iac-web:
    container_name: iac-web
    image: "cloudiac/iac-web:latest"
    ports:
      - 80:80
    restart: always
    depends_on:
      - iac-portal

  mysql:
    container_name: mysql
    image: "mysql:5.7"
    command: [
        "--character-set-server=utf8mb4",
        "--collation-server=utf8mb4_unicode_ci",
        "--sql_mode=STRICT_TRANS_TABLES,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION"
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

#### 5. 编写 .env 文件
文件路径 /usr/yunji/cloudiac/.env，内容如下(**请依据注释修改配置值**):

```bash
# 平台管理员账号，只在初始化启动时进行读取
IAC_ADMIN_EMAIL="admin@example.com"

## 平台管理员密码(必填)，要求长度大于 8 且包含字母、数字、特殊字符
IAC_ADMIN_PASSWORD=""

# 加密密钥配置(必填)
SECRET_KEY=""

# IaC 对外提供服务的地址(必填), 示例: http://cloudiac.example.com
PORTAL_ADDRESS=""

# consul 地址，需要配置为宿主机的内网 ip(必填)，示例: private.host.ip:8500
CONSUL_ADDRESS=""

# mysql 配置(必填)
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_DATABASE=cloudiac
MYSQL_USER=cloudiac
MYSQL_PASSWORD="mysqlpass"

# portal 服务注册信息配置
SERVICE_IP=iac-portal
SERVICE_ID=iac-portal-01
SERVICE_TAGS="iac-portal;portal-01"

## logger 配置
LOG_DEVEL="debug"

# SMTP 配置(该配置只影响邮件通知的发送，不配置也能启动)
SMTP_ADDRESS=smtp.example.com:25
SMTP_USERNAME=user@example.com
SMTP_PASSWORD=""
SMTP_FROM_NAME=IaC
SMTP_FROM=support@example.com

# KAFKA配置(该配置只影响KAFKA消息的发送，不配置也能启动)
KAFKA_TOPIC = "IAC_TASK_REPLY"
KAFKA_GROUP_ID=""
KAFKA_PARTITION=0
## example: KAFKA_BROKERS: ["kafka.example.com:9092", "..."]
KAFKA_BROKERS=[]
KAFKA_SASL_USERNAME=""
KAFKA_SASL_PASSWORD=""

######### 以下为 runner 配置 #############
# runner 服务注册配置
RUNNER_SERVICE_IP=ct-runner
RUNNER_SERVICE_ID=ct-runner-01
RUNNER_SERVICE_TAGS="ct-runner;runner-01"

```

#### 6. 启动docker-compose
```bash
docker-compose up -d

# 前台调试启动：
# dokcker-compose up
```

*至此服务部署完成，如果需要快速创建演示组织请参考文档: [创建演示组织](../demo-org/)*