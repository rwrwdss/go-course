// Команда ci имитирует CI-пайплайн (как GitHub Actions / GitLab CI) для модулей курса.
//
// Использование (из корня курса):
//
//	go run ./tools/ci 01          # прогнать пайплайн модуля 01-...
//	go run ./tools/ci all         # прогнать все модули и вывести сводку
//
// Стадии: gofmt → build → vet → tests (-race) → coverage (если задан порог в ci.cfg).
// Если стадия падает, следующие пропускаются — так же, как в настоящем CI.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	testTimeout  = "60s"
	maxFailLines = 14
)

var useColor = os.Getenv("NO_COLOR") == ""

func paint(code, s string) string {
	if !useColor {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func green(s string) string  { return paint("32", s) }
func red(s string) string    { return paint("31", s) }
func yellow(s string) string { return paint("33", s) }
func gray(s string) string   { return paint("90", s) }
func bold(s string) string   { return paint("1", s) }

type stageResult struct {
	ok      bool
	details []string
}

type stage struct {
	icon string
	name string
	run  func(dir string, cfg moduleConfig) stageResult
}

type moduleConfig struct {
	coverageMin float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Использование: go run ./tools/ci <номер модуля | all>")
		fmt.Println("Пример:        go run ./tools/ci 01")
		os.Exit(2)
	}

	modules, err := findModules()
	if err != nil {
		fmt.Println(red("не удалось прочитать список модулей: " + err.Error()))
		os.Exit(2)
	}

	arg := strings.TrimSuffix(strings.TrimPrefix(os.Args[1], "./"), "/")
	if arg == "all" {
		os.Exit(runAll(modules))
	}

	dir := matchModule(modules, arg)
	if dir == "" {
		fmt.Println(red("модуль не найден: " + arg))
		fmt.Println("Доступные модули:")
		for _, m := range modules {
			fmt.Println("  " + m)
		}
		os.Exit(2)
	}

	if !runPipeline(dir, false) {
		os.Exit(1)
	}
}

func findModules() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`^\d{2}-`)
	var mods []string
	for _, e := range entries {
		if e.IsDir() && re.MatchString(e.Name()) {
			mods = append(mods, e.Name())
		}
	}
	sort.Strings(mods)
	return mods, nil
}

func matchModule(modules []string, arg string) string {
	if len(arg) == 1 {
		arg = "0" + arg
	}
	for _, m := range modules {
		if m == arg || strings.HasPrefix(m, arg+"-") {
			return m
		}
	}
	return ""
}

func runAll(modules []string) int {
	type row struct {
		name string
		ok   bool
	}
	var rows []row
	for _, m := range modules {
		rows = append(rows, row{m, runPipeline(m, true)})
	}
	fmt.Println()
	fmt.Println(bold("Сводка по курсу"))
	passed := 0
	for _, r := range rows {
		mark := red("❌ FAILED")
		if r.ok {
			mark = green("✅ PASSED")
			passed++
		}
		fmt.Printf("  %-34s %s\n", r.name, mark)
	}
	fmt.Printf("\nПройдено модулей: %d/%d\n", passed, len(rows))
	if passed != len(rows) {
		return 1
	}
	return 0
}

func runPipeline(dir string, compact bool) bool {
	cfg := readConfig(dir)
	stages := []stage{
		{"🧹", "gofmt", stageFmt},
		{"🔨", "build", stageBuild},
		{"🔍", "vet", stageVet},
		{"🧪", "tests", stageTests},
	}
	if cfg.coverageMin > 0 {
		stages = append(stages, stage{"📊", "coverage", stageCoverage})
	}

	line := strings.Repeat("═", 56)
	fmt.Println()
	fmt.Println(bold(line))
	fmt.Println(bold("  CI Pipeline · " + dir))
	fmt.Println(bold(line))

	failedAt := ""
	for i, st := range stages {
		prefix := fmt.Sprintf("[%d/%d] %s %-9s", i+1, len(stages), st.icon, st.name)
		if failedAt != "" {
			fmt.Println(gray(prefix + " ⏭  skipped"))
			continue
		}
		start := time.Now()
		res := st.run(dir, cfg)
		elapsed := gray(fmt.Sprintf("(%.1fs)", time.Since(start).Seconds()))
		if res.ok {
			fmt.Println(prefix, green("✅ passed"), elapsed)
		} else {
			fmt.Println(prefix, red("❌ failed"), elapsed)
			failedAt = st.name
		}
		printDetails(res.details, compact && res.ok)
	}

	fmt.Println(bold(line))
	if failedAt != "" {
		fmt.Println(red(bold("  ❌ PIPELINE FAILED")) + " на стадии " + bold(failedAt))
		fmt.Println(gray("  Подсказки — в " + dir + "/README.md"))
		return false
	}
	fmt.Println(green(bold("  ✅ PIPELINE PASSED")) + " — модуль выполнен, можно мёржить 🎉")
	return true
}

func printDetails(lines []string, skip bool) {
	if skip {
		return
	}
	for _, l := range lines {
		fmt.Println("      " + l)
	}
}

func readConfig(dir string) moduleConfig {
	var cfg moduleConfig
	f, err := os.Open(filepath.Join(dir, "ci.cfg"))
	if err != nil {
		return cfg
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		key, val, ok := strings.Cut(strings.TrimSpace(sc.Text()), "=")
		if !ok || strings.HasPrefix(key, "#") {
			continue
		}
		if strings.TrimSpace(key) == "coverage_min" {
			cfg.coverageMin, _ = strconv.ParseFloat(strings.TrimSpace(val), 64)
		}
	}
	return cfg
}

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

func splitLines(s string) []string {
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func stageFmt(dir string, _ moduleConfig) stageResult {
	out, err := run("gofmt", "-l", dir)
	if err != nil {
		return stageResult{false, splitLines(out)}
	}
	files := splitLines(out)
	if len(files) == 0 {
		return stageResult{ok: true}
	}
	details := []string{yellow("Файлы не отформатированы:")}
	for _, f := range files {
		details = append(details, "  "+f)
	}
	details = append(details, "Исправь командой: "+bold("gofmt -w "+dir))
	return stageResult{false, details}
}

func stageBuild(dir string, _ moduleConfig) stageResult {
	out, err := run("go", "build", "-o", os.DevNull, "./"+dir+"/...")
	if err != nil && strings.Contains(out, "-o") && strings.Contains(out, "multiple packages") {
		out, err = run("go", "build", "./"+dir+"/...")
	}
	return stageResult{err == nil, splitLines(out)}
}

func stageVet(dir string, _ moduleConfig) stageResult {
	out, err := run("go", "vet", "./"+dir+"/...")
	return stageResult{err == nil, splitLines(out)}
}

// testEvent — строка из `go test -json`.
type testEvent struct {
	Action  string
	Package string
	Test    string
	Output  string
}

type testInfo struct {
	name   string
	status string
	output []string
}

var lastCoverage = map[string]float64{}

func stageTests(dir string, _ moduleConfig) stageResult {
	cmd := exec.Command("go", "test", "-race", "-count=1", "-timeout", testTimeout, "-cover", "-json", "./"+dir+"/...")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	runErr := cmd.Run()

	tests := map[string]*testInfo{}
	var order []string
	var pkgOutput []string
	covRe := regexp.MustCompile(`coverage: ([\d.]+)% of statements`)
	var totalCov, covCount float64

	sc := bufio.NewScanner(&stdout)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		var ev testEvent
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			pkgOutput = append(pkgOutput, sc.Text())
			continue
		}
		if ev.Test == "" {
			if ev.Action == "output" || ev.Action == "build-output" {
				if m := covRe.FindStringSubmatch(ev.Output); m != nil {
					v, _ := strconv.ParseFloat(m[1], 64)
					totalCov += v
					covCount++
				}
				pkgOutput = append(pkgOutput, strings.TrimRight(ev.Output, "\n"))
			}
			continue
		}
		key := ev.Package + "|" + ev.Test
		ti, ok := tests[key]
		if !ok {
			ti = &testInfo{name: ev.Test}
			tests[key] = ti
			order = append(order, key)
		}
		switch ev.Action {
		case "pass", "fail", "skip":
			ti.status = ev.Action
		case "output":
			ti.output = append(ti.output, strings.TrimRight(ev.Output, "\n"))
		}
	}
	if covCount > 0 {
		lastCoverage[dir] = totalCov / covCount
	}

	var details []string
	passed, skipped, total := 0, 0, 0
	for _, key := range order {
		ti := tests[key]
		depth := strings.Count(ti.name, "/")
		indent := strings.Repeat("   ", depth)
		short := ti.name
		if i := strings.LastIndex(short, "/"); i >= 0 {
			short = short[i+1:]
		}
		short = strings.ReplaceAll(short, "_", " ")
		if depth == 0 {
			total++
		}
		switch ti.status {
		case "pass":
			if depth == 0 {
				passed++
			}
			details = append(details, indent+green("✅ ")+short)
		case "skip":
			if depth == 0 {
				skipped++
			}
			details = append(details, indent+yellow("⏭  ")+short+gray(" (пропущен: TODO)"))
		default:
			details = append(details, indent+red("❌ ")+short)
			if !hasSubtests(tests, key) {
				lines := cleanOutput(ti.output)
				if len(lines) > maxFailLines {
					lines = append(lines[:maxFailLines], "… (вывод обрезан)")
				}
				for _, l := range lines {
					details = append(details, indent+"     "+gray("│ ")+l)
				}
				if containsPanic(ti.output) {
					details = append(details, indent+"     "+yellow("⚠️  Паника/fatal error уронили тестовый бинарник — остальные тесты пакета не запускались."))
				}
			}
		}
	}

	if runErr != nil && total == passed+skipped {
		// Тесты не запустились: ошибка компиляции, паника, гонка, deadlock или таймаут.
		details = append(details, red("Тестовый бинарник упал или не собрался:"))
		for _, l := range append(pkgOutput, splitLines(stderr.String())...) {
			if strings.TrimSpace(l) == "" || strings.HasPrefix(l, "FAIL") || strings.HasPrefix(l, "ok ") {
				continue
			}
			details = append(details, "  "+l)
		}
		if len(details) > 60 {
			details = append(details[:60], gray("  … (вывод обрезан, запусти `go test -v ./"+dir+"/...`)"))
		}
	}

	if total > 0 {
		summary := fmt.Sprintf("Итог: %d/%d тестов пройдено", passed, total)
		if skipped > 0 {
			summary += fmt.Sprintf(", %d пропущено", skipped)
		}
		if passed == total && runErr == nil {
			details = append(details, bold(green(summary)))
		} else {
			details = append(details, bold(red(summary)))
		}
	}
	return stageResult{runErr == nil, details}
}

func hasSubtests(tests map[string]*testInfo, key string) bool {
	for k := range tests {
		if strings.HasPrefix(k, key+"/") {
			return true
		}
	}
	return false
}

func containsPanic(lines []string) bool {
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "panic:") || strings.HasPrefix(t, "fatal error:") {
			return true
		}
	}
	return false
}

func cleanOutput(lines []string) []string {
	var out []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || strings.HasPrefix(t, "=== ") || strings.HasPrefix(t, "--- ") {
			continue
		}
		out = append(out, t)
	}
	return out
}

func stageCoverage(dir string, cfg moduleConfig) stageResult {
	cov, ok := lastCoverage[dir]
	if !ok {
		return stageResult{false, []string{"не удалось определить покрытие"}}
	}
	msg := fmt.Sprintf("Покрытие: %.1f%% (минимум %.0f%%)", cov, cfg.coverageMin)
	if cov < cfg.coverageMin {
		return stageResult{false, []string{red(msg), "Допиши тесты — посмотреть непокрытые строки:",
			"  go test -coverprofile=cover.out ./" + dir + "/ && go tool cover -html=cover.out"}}
	}
	return stageResult{true, []string{green(msg)}}
}
