CloudIaC
================
Cloud Infrastructure as code

## 编译
1. 配置 go env
```
export GO111MODULE="on" 
export GOPROXY="https://goproxy.io"
```

2. 执行编译
```
make all|portal|runner
```
可执行文件将生成到 `./builds/` 目录


## 配置
1. 连接 mysql, 创建 db
```sql
create database iac charset utf8mb4;
```

2. 拷贝配置模板
```
cp configs/config.yml.sample config.yml
cp configs/dotenv.sample .env
```

3. 配置 `.env`

主要进行 db 信息配置

## 运行
```
make run
```

