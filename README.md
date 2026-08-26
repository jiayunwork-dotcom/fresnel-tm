# fresnel-tm

fresnel-tm 是一个多层介质膜特征矩阵核算工具：给定入射介质、各层折射率/厚度、衬底以及入射波长与角度，它算出反射率 R、透射率 T 与吸收率 A，并在波长区间上扫描出 R(λ)/T(λ)/A(λ) 光谱。求解器用 2×2 复特征矩阵连乘实现（每层 M_j = [[cos δ, i sin δ/η], [i η sin δ, cos δ]]，δ = 2π·n·d·cosθ/λ），s/p 偏振分别使用不同的导纳规则 η = n·cosθ / η = n/cosθ。无吸收膜系满足 R + T = 1；含吸收介质（复折射率）时 R + A + T = 1 且 A ≥ 0；层数为 0 时退化为单界面 Fresnel 反射。工具自带一个 Web 控制台，加载示例后即可计算并绘制光谱，所有数值都来自后端求解。

## 用法

启动 Web 控制台（页面与 /api 同进程，示例与页面编译进二进制）：

```bash
go run . -http :8080
```

打开 <http://localhost:8080>：点击示例「ar-quarter」加载预置算例（玻璃上 λ/4 的 MgF2 减反膜），点「计算」看设计波长 550 nm 处 R ≈ 1.41%（裸玻璃约 4.0%），点「扫描光谱」看 R(λ) 折线与减反凹坑位置。

命令行直接核算一个示例文件（不启动 HTTP）：

```bash
go run . -solve example/ar-quarter.json
go run . -solve example/absorber-stack.json
```

## API

- `POST /api/stack` — 膜系 + 波长 + 入射角 + 偏振 → `{reflection, transmission, absorption, energy_sum, bare_reflection, ...}`
- `POST /api/spectrum` — 膜系 + 波长区间 + 采样点数 → `{points: [{wavelength_nm, reflection, transmission, absorption}], min_reflection, ...}`
- `GET /api/examples` / `GET /api/examples/{name}` — 内置算例列表与内容

请求体与 `example/*.json` 同构：

```json
{
  "incident": {"index": 1.0},
  "layers": [{"index": 1.38, "thickness_nm": 99.64}],
  "substrate": {"index": 1.5},
  "wavelength_nm": 550,
  "angle_deg": 0,
  "polarization": "average"
}
```

非法输入（波长 ≤ 0、层厚为负、n ≤ 0、入射角 ≥ 90°、未知字段等）一律返回 HTTP 400 与 `{"error": "..."}`，错误文案直接可在页面看到。光谱请求的字段为 `wavelength_min_nm` / `wavelength_max_nm` / `points`。

## 约定与边界

- 长度单位一律为纳米（nm），角度为度；入射角范围 [0°, 90°)，波长必须 > 0。
- 入射介质必须无吸收；衬底与镀层可以为吸收介质（如金属 Cr）。
- 减反交叉规则：把 λ/4 层光程改成 λ/2，设计波长处的减反失效、R 回升至裸基板水平；增大入射角接近掠射时 R 升高；斜入射下减反凹坑随 cosθ 向短波蓝移。
- 本工具只做平面波在平行平板膜系中的 R/T/A，不做偏振级联、波片相位或色散材料库。

## 构建与测试

```bash
go build ./...
go test ./...
```

代码只依赖 Go 标准库，模块名为 `fresnel-tm`，Go 1.21。
