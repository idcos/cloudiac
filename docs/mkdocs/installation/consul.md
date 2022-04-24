## 容器部署开启acl和tls

### 修改 docker-compose.yml 文件

文件路径 /usr/yunji/cloudiac/docker-compose.yml，内容如下:

```yaml
# auto-replace-from: /usr/yunji/cloudiac/docker-compose.yml
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
        source: /usr/yunji/cloudiac
        target: /consul/config
    ports:
      - "8500:8500"
    command: >
      consul agent -server -bootstrap-expect=1 -ui -bind=0.0.0.0
      -client=0.0.0.0 -enable-script-checks=true -data-dir=/consul/data -config-dir=/consul/config
    restart: always

```

> 配置说明

!!! Caution
    `docker-compose.yaml` 中 **consul** 新增以下配置，其他配置可根据需要修改

    - CONSUL_HTTP_SSL_VERIFY=false: 私有化部署SSL证书不验证
    - CONSUL_HTTP_SSL=true: 启动https URI方案和http api的SSl连接
    - -config-dir=/consul/config: 容器启动Command新增挂载指定配置目录
    - /usr/yunji/cloudiac:/consul/config:新增挂载目录 


### 开启acl配置

文件路径 /usr/yunji/cloudiac/acl.hcl,内容如下:

```yaml
# auto-replace-from: /usr/yunji/cloudiac/acl.hcl
cat >> /usr/yunji/cloudiac/acl.hcl <<EOF
acl = {
  enabled = true
  default_policy = "deny"
  enable_token_persistence = true
}
EOF
```

### 开启tls配置文件

文件路径 /usr/yunji/cloudiac/tls.json,内容如下:

- 创建证书并添加tls.json配置

> 证书名称固定 ca.pem,client.key,client.pem

```bash
cd /usr/yunji/cloudiac/

 #生成根证书key
openssl genrsa -out ca.key 2048
#生成根证书密钥
openssl req -new -x509 -days 7200 -key ca.key   -out ca.pem

#生成客户端私钥
openssl genrsa -out client.key 2048

#生成的客户端的CSR
openssl req -new -key client.key  -out client.csr

# 创建宿主机机对应的签名证书  IP为当前部署环境的主机ip,不可用127.0.0.1或者localhost
echo subjectAltName = IP:xx.xx.xx.xx > extfile.cnf

#客户端自签名的证书
openssl x509 -req -days 365 -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
   -out client.pem -extfile extfile.cnf
   
# 新增consul配置tls.json
cat >> /usr/yunji/cloudiac/tls.json <<EOF
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

### 重启consul

```bash
docker-compose up consul
```
> 默认为前台启动，以便于排查问题，在确定服务正常后可以改为后台启动：`dokcker-compose up -d consul`。



### 配置consul acl的token
```bash
# 进入容器
docker exec -it consul sh

# 生成token,保存好生成的SecretID
consul acl bootstrap

#退出容器

# 加入SecretID作为token加入acl.hcl配置
cat > /usr/yunji/cloudiac/acl.hcl <<EOF
acl = {
  enabled = true
  default_policy = "deny"
  enable_token_persistence = true
  tokens {
    master = "a0419d88-cd14-f96f-e144-a02a0f03f683" 
  }
}
EOF
```
!!! Info
    consul acl bootstrap执行结果如下,SecretID为所需要的token
    ```bash
    # consul acl bootstrap
    AccessorID:       af48d2cf-690d-eafe-5e5a-40e3239efa9e
    SecretID:         a0419d88-cd14-f96f-e144-a02a0f03f683
    Description:      Bootstrap Token (Global Management)
    Local:            false
    Create Time:      2022-04-14 09:00:05.914372 +0000 UTC
    Policies:
    00000000-0000-0000-0000-000000000001 - global-management
    ```


---

## 二进制部署开启acl和tls
### consul开启acl

```bash
# 新增consul配置acl.hcl
cat >> /etc/consul.d/acl.hcl <<EOF
acl = {
  enabled = true
  default_policy = "deny"
  enable_token_persistence = true
}
EOF

# 重启consul
systemctl restart consul

# 生成token
consul acl bootstrap

# 加入SecretID作为token加入acl.hcl配置
cat > /etc/consul.d/acl.hcl <<EOF
acl = {
  enabled = true
  default_policy = "deny"
  enable_token_persistence = true
  tokens {
    master = "a0419d88-cd14-f96f-e144-a02a0f03f683" 
  }
}
EOF
```
!!! Info
    consul acl bootstrap执行结果如下,SecretID为所需要的token
    ```bash
    # consul acl bootstrap
    AccessorID:       af48d2cf-690d-eafe-5e5a-40e3239efa9e
    SecretID:         a0419d88-cd14-f96f-e144-a02a0f03f683
    Description:      Bootstrap Token (Global Management)
    Local:            false
    Create Time:      2022-04-14 09:00:05.914372 +0000 UTC
    Policies:
    00000000-0000-0000-0000-000000000001 - global-management
    ```

### consul开启tls

> 证书名称固定 ca.pem,client.key,client.pem

```bash
cd /usr/yunji/cloudiac 
 
 #生成根证书key
openssl genrsa -out ca.key 2048
#生成根证书密钥
openssl req -new -x509 -days 7200 -key ca.key   -out ca.pem

#生成客户端私钥
openssl genrsa -out client.key 2048

#生成的客户端的CSR
openssl req -new -key client.key  -out client.csr

# 创建宿主机机对应的签名证书  IP为当前部署环境的主机ip,不可用127.0.0.1或者localhost
echo subjectAltName = IP:xx.xx.xx.xx > extfile.cnf

#客户端自签名的证书
openssl x509 -req -days 365 -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
   -out client.pem -extfile extfile.cnf
   
# 新增consul配置tls.json
cat >> /etc/consul.d/tls.json <<EOF
{
  "verify_incoming": false,
  "verify_incoming_rpc": true,
  "ports": {
    "http": -1,
    "https": 8500
  },
  "ca_file": "/usr/yunji/cloudiac/ca.pem",
  "cert_file": "/usr/yunji/cloudiac/client.pem",
  "key_file": "/usr/yunji/cloudiac/client.key"
}
EOF

# 新增环境变量
cat >> /etc/profile <<EOF
export CONSUL_HTTP_SSL=true
export CONSUL_HTTP_SSL_VERIFY=false
EOF

#环境变量生效
source /etc/profile

# 重启consul
systemctl restart consul
```

!!! info
    环境变量说明

    - CONSUL_HTTP_SSL_VERIFY=false: 私有化部署SSL证书不验证
    - CONSUL_HTTP_SSL=true: 启动https URI方案和http api的SSl连接
