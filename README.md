## 简介

Backup 是一个用Go语言编写的自动化备份工具，用于备份文件。该工具会将指定目录（/app/docker）下的文件压缩成tar.gz格式，并定期上传到阿里云OSS（对象存储服务）中。

## 特性

- 自动压缩指定目录下的文件
- 定时备份（默认每天凌晨1点执行）
- 使用阿里云OSS存储备份文件
- 异步日志记录
- 使用Docker容器运行，确保环境一致性

## 环境要求

- Docker
- 阿里云OSS账户

## 使用说明

### 1. 环境变量配置

在运行Docker容器之前，请确保设置以下环境变量：

- `ENDPOINT`: 阿里云OSS的endpoint
- `ACCESS_KEY_ID`: 阿里云账户的Access Key ID
- `ACCESS_KEY_SECRET`: 阿里云账户的Access Key Secret
- `BUCKET_NAME`: 用于存储备份文件的OSS Bucket名称

### 2. 构建Docker镜像

在项目根目录下运行以下命令构建Docker镜像：

```bash
docker build -t docker-backup-tool .
```

### 3. 运行Docker容器

使用以下命令运行Docker容器：

```bash
docker run -d \
  -e ENDPOINT=<your-oss-endpoint> \
  -e ACCESS_KEY_ID=<your-access-key-id> \
  -e ACCESS_KEY_SECRET=<your-access-key-secret> \
  -e BUCKET_NAME=<your-bucket-name> \
  -v /path/to/docker/files:/app/docker \
  -v /path/to/logs:/app/logs \
  docker-backup-tool
```

请将尖括号 `<>` 中的内容替换为您的实际配置。

### 4. 查看日志

备份操作的日志将被写入到容器内的 `/app/logs/backup.log` 文件中。您可以通过以下命令查看日志：

```bash
docker exec -it <container_id> cat /app/logs/backup.log
```

## 注意事项

- 默认情况下，备份操作每天凌晨1点执行一次。如果需要修改备份频率，请修改 `main.go` 文件中的相关代码。
- 请确保为阿里云OSS账户配置了正确的访问权限。
- 备份文件将被存储在指定的OSS Bucket的 `backups/` 目录下，文件名格式为 `docker_YYYYMMDD_HHMMSS.tar.gz`。

## 贡献

欢迎提交问题报告和拉取请求，以帮助改进这个项目。

## 许可证

MIT
