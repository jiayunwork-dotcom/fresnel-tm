# fresnel-tm：Go 多层介质膜特征矩阵 Web 服务（R/T/A 光谱 + 前端控制台）

给定入射介质、各层折射率/厚度与衬底，按波长与角度算出反射率、透射率与吸收率；提供 `/api/stack`、`/api/spectrum` 与嵌入网页。

## 构建 / 运行 / 测试

```text
go build ./...
./fresnel-tm -http :8080
curl -s http://127.0.0.1:8080/api/examples
go run . -solve example/ar-quarter.json
go test ./...
```

## 评测镜像

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -d -P --name fresnel-tm-b14 <image-name>:latest
curl -s http://127.0.0.1:$(docker port fresnel-tm-b14 8080 | cut -d: -f2)/api/examples
docker rm -f fresnel-tm-b14
```
