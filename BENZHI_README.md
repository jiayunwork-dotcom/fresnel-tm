# fresnel-tm：多层介质膜特征矩阵核算工具

给定入射介质、各层折射率/厚度与衬底，按入射波长与角度算出反射率 R、透射率 T、吸收率 A，并在波长区间上扫描 R(λ)/T(λ)/A(λ) 光谱；附带 Web 控制台（Go 同进程提供页面与 /api）。

## 构建 / 运行 / 测试

```text
go build ./...
go run . -http :8080       # Web 控制台：加载 example/ar-quarter.json 计算并画光谱
go run . -solve example/ar-quarter.json   # 命令行核算示例（R ≈ 1.41%，裸玻璃 ≈ 4.0%）
go test ./...
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

两种架构都要构建并进容器验证：

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```
