package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ParallaxRoot/argusenum/internal/config"
	"github.com/ParallaxRoot/argusenum/internal/core"
	"github.com/ParallaxRoot/argusenum/internal/logger"
)

func main() {
	printBanner()

	domain := flag.String("d", "", "Single domain to enumerate (e.g. example.com)")
	listPath := flag.String("list", "", "Path to file with one domain per line")
	outputPath := flag.String("o", "results.json", "Output JSON file path")

	passiveOnly := flag.Bool("passive", false, "Run only passive enumeration")
	activeOnly := flag.Bool("active", flase, "Run only active enumeration (bruteforce/permutations)")
	resolversFile := flag.String("resolvers", "", "Custom resolvers file (optional)")
	threads := flag.Int("threads", 64, "Number of worker threads for DNS/HTTP")

	flag.Parse()

	if *domain == "" && *listPath == "" {
		fmt.Fprintln(os.Stderr, "[!] You must pass -d <domain> or -list <file>")
		flag.Usage()
		os.Exit(1)
	}

	var domains []string
	if *domain != "" {
		domains = append(domains, *domain)
	}

	if *listPath != "" {
		listDomains, err := loadDomainsFromFile(*listPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] Failed to read list file: %v\n", err)
			os.Exit(1)
		}
		domains = append(domains, listDomains...)
	}

	if len(domains) == 0 {
		fmt.Fprintln(os.Stderr, "[!] No valid domains provided")
		os.Exit(1)
	}

	cfg := config.Config{
		Domains:      domains,
		Output:       *outputPath,
		PassiveOnly:  *passiveOnly,
		ActiveOnly:   *activeOnly,
		ResolversFile: *resolversFile,
		Threads:      *threads,
	}

	log := logger.New()

	engine := core.NewEngine(cfg, log)

	if err := engine.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[!] Execution error: %v\n", err)
		os.Exit(1)
	}
}

func printBanner() {
	fmt.Println(`
   ╔══════════════════════════════════════════════╗
   ║              ArgusEnum v0.1                 ║
   ║     Next-gen subdomain enumerator           ║
   ║     by ParallaxRoot (HackerOne)             ║
   ╚══════════════════════════════════════════════╝
`)
}

func loadDomainsFromFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var domains []string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domains = append(domains, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return domains, nil
}