# GO web sockets example

### 开启服务端
```bash
./ws -p=8000
```

### 测试客户端
```bash
cd wstest && go run ws_client.go
```


### 语音盒子
192.168.55.1 1883

$queue/device/#


链接: https://pan.baidu.com/s/1gsBAxbfrXGjptEtYxlBCiw  密码: ffak

使用方法：
 1. 盒子帐号密码：devops，devops
 2. 使用scp将镜像文件传入盒子
 3. 加载镜像
  docker load < arm_asr_v20210218_streaming.tar.gz
 4. 由于内存不足，需关闭之前已部署的容器服务
  docker ps 查看容器id
  docker rm -f 容器id
 5. 启动asr
  docker run -d -p 8001:8001 -e MASTER_PORT=8001 -e NUM_ASR_WORKER=2 zijingtaoli-docker.pkg.coding.net/asr-server/image/asr:v20210218_arm  supervisord -c supervisord.conf -n
  
 6. 启动过程约30s，内存会上涨至2.5-2.7G