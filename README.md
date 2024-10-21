# minio-compress

## 使用说明

1. 设置环境变量

```shell
export MINIO_ENDPOINT=10.160.56.234:9000
export MINIO_ACCESS_KEY_ID=property
export MINIO_SECRET_ACCESS_KEY=Property@123
export MINIO_BUCKET_NAME=inzone-property
export MINIO_FILE_PREFIX=2024/01/01
export PROCESS_NUM=20
```

> MINIO_FILE_PREFIX:需要处理的文件前缀  2024/01/01:表示处理2024年1月1日的文件  2024/01:表示处理2024年1月的文件  
> PROCESS_NUM:并发处理的图片数量,根据机器配置填写

2. 运行

```shell
chmod +x compress
./compress
```
