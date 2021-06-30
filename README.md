CloudIaC
================
Cloud Infrastructure as code

## 编译
1. 配置 go env
```
go env -w GO111MODULE="on" 
go env -w GOPROXY="https://goproxy.io,direct"
```

2. 执行编译
```
make all 
# or 'make portal' 'make runner'
```
可执行文件将生成到 `./targets/` 目录

## 本地调试运行
1. 拷贝配置模板
```
cp configs/config-portal.yml.sample config-portal.yml
cp configs/config-runner.yml.sample config-runner.yml
cp configs/dotenv.sample .env
```

2. 启动 mysql 服务, 创建 db
```sql
create database iac charset utf8mb4;
```

3. 启动 consul 服务

4. 配置 
编辑 `.env` 和 `config-runner.yml`

5. 启动
打开两个终端，分别运行:
```
make run-portal
```

```
make run-runner
```

