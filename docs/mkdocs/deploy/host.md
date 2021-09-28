# 手动部署

## 环境依赖
1. 操作系统：Linux x86_64 架构
1. 中间件：MySQL 5.7、Consul

## 后端部署
直接在物理机或者虚拟机上部署 IaC 服务。
限定条件：

- 单机部署
- 全新机器初始安装
- 使用 centos 7 系统
- 部署目录为 /usr/yunji/cloudiac


以下部署过程全部使用 **root** 操作。
#### 1. 下载并解压安装包

目前 cloudiac 分为三个包:

- 主程序: cloudiac_${VERSION}.tar.gz
- 预置样板间: cloudiac-repos_${VERSION}.tar.gz
- 预置 proviers: cloudiac-providers_${VERSION}.tar.gz 


```
VERSION=v0.5.1
mkdir -p /usr/yunji/cloudiac && \
cd /usr/yunji/cloudiac && \
for PACK in cloudiac cloudiac-repos cloudiac-providers; do
  curl -sL https://github.com/idcos/cloudiac/releases/download/${VERSION}/${PACK}_${VERSION}.tar.gz -o ${PACK}_${VERSION}.tar.gz && \
  tar -xf ${PACK}_${VERSION}.tar.gz
done
```

**目前只支持部署到: `/usr/yunji/cloudiac` 目录**


#### 2. 安装并启动 Docker
```bash
curl -fsSL https://get.docker.com | bash -s docker

systemctl enable docker
systemctl start docker
```

#### 3. 安装并启动 Mysql
```bash
yum install -y https://repo.mysql.com/mysql57-community-release-el7.rpm
yum install -y mysql-server

systemctl enable mysqld
systemctl start mysqld
```

安装完成后直接使用 `mysql -uroot` 即可连接到数据库，然后新创建一个账号用于 cloudiac。
> 根据版本和系统的不同，安装后的初始密码可能在日志中打印或者写入到了 /etc/mysql/xxx 配置中，请根据实际情况获取默认密码。

#### 4. 安装并启动 Consul

```bash
yum install -y yum-utils
yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo
yum -y install consul

## 修改 consul 配置
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

以上配置仅用于单机测试时使用，完整的 consul 集群部署请参考官方文档:    
https://learn.hashicorp.com/tutorials/consul/get-started-install

#### 5. 编辑配置文件

###### 拷贝示例配置文件

```bash
mv config-portal.yml.sample config-portal.yml
mv config-runner.yml.sample config-runner.yml
mv dotenv.sample .env
mv demo-conf.yml.sample demo-conf.yml
```

###### 修改配置 

编辑 .env 文件，主要修改如下内容:

```bash
# 管理员账号密码，只在初始化启动时进行读取
IAC_ADMIN_EMAIL=admin@example.com
IAC_ADMIN_PASSWORD=

# IaC 对外提供服务的地址
PORTAL_ADDRESS=http://public.host.ip

# mysql 配置
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_DATABASE=iac
MYSQL_USER=root
MYSQL_PASSWORD=

# consul 配置
CONSUL_ADDRESS=private.host.ip:8500

# portal 服务注册信息配置
SERVICE_IP=private.host.ip
SERVICE_ID=IaC-Portal-01
SERVICE_TAGS=iac-portal;portal-01
```

###### 编辑 config-runner.yml 文件，主要修改如下内容:

```yaml
consul:
  address: "consul.example.com:8500"
  // runner id 全局唯一
  id: "CT-Runner-01"
  // runner 的地址，该地址需要确保 portal 和 consul 能正常访问
  ip: "runner01.example.com"
  tags: "ct-runner;runner-01"
```

#### 6. 初始化 Mysql

```sql
-- 连接到 mysql 执行以下命令创建 db，默认配置的 db 名称为 iac
create database iac charset utf8mb4;
```

#### 7. 安装 IaC 服务

```shell
cp iac-portal.service ct-runner.service /etc/systemd/system/
systemctl enable iac-portal ct-runner
```

#### 8. 启动 IaC 服务

```shell
## 启动服务
systemctl start iac-portal ct-runner

## 确定服务状态
systemctl status -l iac-portal ct-runner
```

#### 9. 接取 ct-worker 镜像

ct-worker 是执行部署任务的容器镜像，需要提前 pull 到本地, 命令如下(该操作可以后台进行):
```
docker pull cloudiac/ct-worker
```

#### 10. 初始化演示组织（可选步骤）

复制云商密钥对私钥并存储在本地，比如，存储为 /usr/yunji/cloudiac/var/private_key 文件

###### 编辑 demo-conf.yml 文件，主要修改如下内容:

```yaml
  # 云商密钥对私钥文件路径
  key:
    name: 示例密钥
    key_file: ./var/private_key
  # 云商访问密钥，这里以阿里云为例
  variables:
    - name: ALICLOUD_ACCESS_KEY
      type: environment
      value: xxxxxxxxxx
      description: 阿里云访问密钥ID
      sensitive: true
    - name: ALICLOUD_SECRET_KEY
      type: environment
      value: xxxxxxxxxx
      description: 阿里云访问密钥
      sensitive: true
```

###### 初始化并重启服务

```shell
## 初始化演示组织
cd /usr/yunji/cloudiac && ./iac-tool init-demo demo-conf.yml

## 重启 portal 服务
systemctl restart iac-portal
```

** 注意演示组织只能创建一次，如需重新初始化，需手动删除相关数据 **

至此后端部署完成。


## 前端部署
#### 1. 下载前端部署包并解压
```
VERSION=v0.5.1
mkdir -p /usr/yunji/cloudiac-web && \
cd /usr/yunji/cloudiac-web && \
curl -sL https://github.com/idcos/cloudiac-web/releases/download/${VERSION}/cloudiac-web_${VERSION}.tar.gz -o cloudiac-web_${VERSION}.tar.gz && \
tar -xf cloudiac-web_${VERSION}.tar.gz
```

#### 2. 安装 nginx
```bash
yum install -y nginx
```

#### 3. 配置 nginx
配置 nginx 实现前端静态文件访问，并代理后端接口。
nginx 示例配置:
```
server {
  listen 80;
  server_name _ default;

  location / {
    try_files $uri $uri/ /index.html /index.htm =404;
    root /usr/yunji/cloudiac-web;
    index  index.html index.htm;
  }

  location = /login {
    rewrite ^/login /login.html last;
  }

  location /api/v1/ {
    proxy_buffering off;
    proxy_cache off;

    proxy_read_timeout 1800;
    proxy_pass http://iac-portal:9030;
  }

  location /repos/ {
    proxy_pass http://iac-portal:9030;
  }
}
```

- 其中 `iac-portal` 需要替换为后端 portal 服务的 ip

配置后重启 nginx，完成前端部署。

