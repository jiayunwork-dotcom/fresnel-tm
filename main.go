package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"fresnel-tm/internal/model"
	"fresnel-tm/internal/optics"
	"fresnel-tm/internal/server"
)

//go:embed web example
var assets embed.FS

func main() {
	httpAddr := flag.String("http", "", "HTTP listen address for the web console and /api (e.g. :8080)")
	solveFile := flag.String("solve", "", "path to an example JSON; print its design-wavelength solve and exit")
	flag.Parse()

	if *solveFile != "" {
		runSolve(*solveFile)
		return
	}

	if *httpAddr == "" {
		fmt.Fprintln(os.Stderr, "用法：fresnel-tm -http :8080 或 fresnel-tm -solve example/ar-quarter.json")
		os.Exit(2)
	}

	srv := server.NewServer(assets)
	fmt.Fprintf(os.Stderr, "fresnel-tm: listening on %s\n", *httpAddr)
	if err := srv.ListenAndServe(*httpAddr); err != nil {
		fmt.Fprintf(os.Stderr, "fresnel-tm: server error: %v\n", err)
		os.Exit(1)
	}
}

func runSolve(path string) {
	ex, err := model.LoadExampleFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fresnel-tm: %v\n", err)
		os.Exit(1)
	}
	wl := ex.WavelengthNm
	if wl == 0 && ex.DesignWavelengthNm > 0 {
		wl = ex.DesignWavelengthNm
	}
	req := ex.StackRequest
	req.WavelengthNm = wl

	pol, err := req.ResolvedPolarization()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fresnel-tm: %v\n", err)
		os.Exit(1)
	}
	res, err := optics.Solve(req.Stack(), req.Incidence, pol)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fresnel-tm: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("示例：%s\n", ex.Name)
	if ex.Description != "" {
		fmt.Printf("说明：%s\n", ex.Description)
	}
	fmt.Printf("波长 %v nm · 入射角 %v° · 偏振 %s\n", res.WavelengthNm, res.AngleDeg, res.Polarization)
	fmt.Printf("反射率 R = %.6f\n", res.Reflection)
	fmt.Printf("透射率 T = %.6f\n", res.Transmission)
	fmt.Printf("吸收率 A = %.6f\n", res.Absorption)
	fmt.Printf("能量和 R+A+T = %.9f（损失耗时应为 1）\n", res.EnergySum)
	fmt.Printf("裸界面参考：R = %.6f, T = %.6f\n", res.BareReflection, res.BareTransmission)
	if res.Absorption > 1e-9 {
		fmt.Println("结论：膜系有吸收，R + A + T = 1 成立")
	} else if res.Reflection < res.BareReflection {
		fmt.Println("结论：镀膜后反射率低于裸基板（减反生效）")
	} else {
		fmt.Println("结论：镀膜后反射率不低于裸基板（该波长非减反设计点）")
	}
}
