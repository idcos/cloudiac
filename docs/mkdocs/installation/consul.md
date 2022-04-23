###  consul开启acl和tls配置

#### 修改docker-compose.yaml

修改docker-compose.yaml,替换consul部分

文件路径 /usr/yunji/cloudiac/docker-compose.yml，内容如下:
```yaml
# auto-replace-from: docker/docker-compose.yml
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
    `docker-compose.yaml` 中新增以下配置，其他配置可根据需要修改,如果不需要开启tls可以不加环境变量

    - CONSUL_HTTP_SSL_VERIFY=false: 私有化部署SSL证书不验证
    - CONSUL_HTTP_SSL=true: 启动https URI方案和http api的SSl连接
    - -config-dir=/consul/config: Command新增挂载指定配置目录
    - /usr/yunji/cloudiac:/consul/config:新增挂载目录 


!!! Info
    以下配置开启请在替换docker-compose.yaml中consul部分,并执行docker-compose up以后进行。

#### 开启consul配置acl
```bash
# 新增consul配置acl.hcl
cat >> /usr/yunji/cloudiac/acl.hcl <<EOF
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
cat > /usr/yunji/cloudiac/acl.hcl <<EOF
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

#### 开启consul配置tls访问
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

# 创建宿舍机对应的签名证书  IP宿主机
echo subjectAltName = IP:127.0.0.1 > extfile.cnf

#客户端自签名的证书
openssl x509 -req -days 365 -in client.csr -CA ca.pem -CAkey ca.key -CAcreateserial \
   -out client.pem -extfile extfile.cnf
   
# 新增consul配置tls.json
cat >> /usr/yunji/cloudiac <<EOF
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