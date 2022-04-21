# 手动部署

## 环境依赖

1. 操作系统：Centos 7 及以上版本
2. 中间件：MySQL 8.0、Consul

!!! Info
    理论上任意 Linux 发行版本都可以支持部署，但该文档仅以 Centos 系统为例演示部署过程。

## 后端部署

直接在物理机或者虚拟机上部署 IaC 服务。 该文档的环境限定：

- 单机部署
- 全新机器初始安装
- 部署目录为 /usr/yunji/cloudiac

以下部署过程全部使用 **root** 操作。

### 1. 下载并解压安装包

目前 cloudiac 分为三个包:

- 主程序: cloudiac_${VERSION}.tar.gz
- 预置样板间: cloudiac-repos_${VERSION}.tar.gz
- 预置 proviers: cloudiac-providers_${VERSION}.tar.gz

```
VERSION=v0.9.1
mkdir -p /usr/yunji/cloudiac && \
cd /usr/yunji/cloudiac && \
for PACK in cloudiac cloudiac-repos cloudiac-providers; do
  curl -sL https://github.com/idcos/cloudiac/releases/download/${VERSION}/${PACK}_${VERSION}.tar.gz -o ${PACK}_${VERSION}.tar.gz && \
  tar -xf ${PACK}_${VERSION}.tar.gz
done
```

!!! Caution
    **部署目录必须为 /usr/yunji/cloudiac，部署到其他目录将无法执行环境部署任务。**

### 2. 安装并启动 Docker

```bash
curl -fsSL https://get.docker.com | bash -s docker

systemctl enable docker
systemctl start docker
```

### 3. 安装并启动 Mysql

```bash
yum install -y https://repo.mysql.com/mysql57-community-release-el7.rpm
yum install -y mysql-server

systemctl enable mysqld
systemctl start mysqld
```

安装完成后直接使用 `mysql -uroot` 即可连接到数据库，然后新创建一个账号用于 cloudiac。

> 根据版本和系统的不同，安装后的初始密码可能在日志中打印或者写入到了 /etc/mysql/xxx 配置中，请根据实际情况获取默认密码。

### 4. 安装并启动 Consul

```bash
yum install -y yum-utils
yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo
yum -y install consul

# 修改 consul 配置
cat >> /etc/consul.d/consul.hcl <<EOF
ui = true
server = true
bind_addr = "0.0.0.0"
client_addr = "0.0.0.0"
bootstrap_expect = 1
enable_script_checks = true
EOF

systemctl enable consul
systemctl start consul
```

- consul开启acl

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
    master = "5ecc86f0-fa68-6ddc-e848-ef382d7737ec" #SecretID
  }
}
EOF
```

- consul开启tls

> 证书名称固定 ca.pem,client.key,client.pem

```bash
mkdir -p /usr/yunji/tls && \
cd /usr/yunji/tls &&
 
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
cat >> /etc/consul.d/tls.json <<EOF
{
  "verify_incoming": false,
  "verify_incoming_rpc": true,
  "ports": {
    "http": -1,
    "https": 8500
  },
  "ca_file": "/usr/yunji/tls/ca.pem",
  "cert_file": "/usr/yunji/tls/client.pem",
  "key_file": "/usr/yunji/tls/client.key"
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



以上配置仅用于单机测试时使用，完整的 consul 集群部署请参考官方文档:
https://learn.hashicorp.com/tutorials/consul/get-started-install

### 5. 编辑配置文件

- 拷贝示例配置文件

```bash
mv config-portal.yml.sample config-portal.yml
mv config-runner.yml.sample config-runner.yml
mv dotenv.sample .env
mv demo-conf.yml.sample demo-conf.yml
```

- 编辑 .env 文件，依据注释修改配置。

!!! Caution
    `.env` 中以下配置为**必填项**，其他配置可根据需要修改：

    - IAC_ADMIN_PASSWORD: 初始的平台管理员密码
    - SECRET_KEY: 数据加密存储时使用的密钥
    - PORTAL_ADDRESS: 对外地址服务的地址
    - CONSUL_ADDRESS: consul 服务地址，配置为部署机内网 ip:8500 端口即可

!!! Info
    通过 `.env` 可以实现大部分配置的修改，更多配置项可查看 config-portal.yml 和 config-runner.yml。


### 6. 初始化 Mysql

```sql
-- 连接到 mysql 执行以下命令创建 db，默认配置的 db 名称为 iac
create database iac charset utf8mb4;
```

### 7. 安装 IaC 服务

```shell
cp iac-portal.service ct-runner.service /etc/systemd/system/
systemctl enable iac-portal ct-runner
```

### 8. 启动 IaC 服务

```shell
# 启动服务
systemctl start iac-portal ct-runner

# 确定服务状态
systemctl status -l iac-portal ct-runner
```

### 9. 拉取 ct-worker 镜像

ct-worker 是执行部署任务的容器镜像，需要 pull 到本地:

```
docker pull cloudiac/ct-worker
```

该操作可以后台进行，保证在执行环境部署前镜像 pull 到本地即可。

## 前端部署

### 1. 下载前端部署包并解压

```
VERSION=v0.9.1
mkdir -p /usr/yunji/cloudiac-web && \
cd /usr/yunji/cloudiac-web && \
curl -sL https://github.com/idcos/cloudiac-web/releases/download/${VERSION}/cloudiac-web_${VERSION}.tar.gz -o cloudiac-web_${VERSION}.tar.gz && \
tar -xf cloudiac-web_${VERSION}.tar.gz
```

### 2. 安装 nginx

```bash
yum install -y nginx
```

### 3. 配置 nginx

配置 nginx 实现前端静态文件访问，并代理后端接口。 nginx 示例配置:

```
server {
  listen 80;
  server_name _ default;

  gzip  on;
  gzip_min_length  1k;
  gzip_buffers 4 16k;
  gzip_http_version 1.1;
  gzip_comp_level 9;
  gzip_types text/plain application/x-javascript text/css application/xml text/javascript \
    application/x-httpd-php application/javascript application/json;
  gzip_disable "MSIE [1-6]\.";
  gzip_vary on;

  location / {
    try_files $uri $uri/ /index.html /index.htm =404;
    root /usr/nginx/cloudiac-web;
    index  index.html index.htm;
  }

  location = /login {
    rewrite ^/login /login.html last;
  }

  location /api/v1/ {
    proxy_buffering off;
    proxy_cache off;

    proxy_read_timeout 1800;
    proxy_pass http://127.0.0.1:9030;
  }

  location /repos/ {
    proxy_pass http://127.0.0.1:9030;
  }
}
```

配置后重启 nginx，完成前端部署。


## 部署完成
至此服务部署完成，访问 http://${PORTAL_ADDRESS} 进行登陆。

默认的用户名为 admin@example.com (即 IAC_ADMIN_EMAIL)，密码为 `.env` 中配置的 `IAC_ADMIN_PASSWORD`。
