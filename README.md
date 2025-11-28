# ArgusEnum

Next-gen subdomain enumerator by **ParallaxRoot** (HackerOne).

Objetivo: ser mais inteligente que subfinder + amass juntos, com:

- Coleta passiva agressiva (crt.sh, Censys, VT, etc.)
- Bruteforce e permutações guiadas por padrões
- Resolver DNS paralelo
- Checagem HTTP com fingerprint
- Correlação e tagging inteligente
- Futuro: ML para prever subdomínios prováveis

## Uso (versão inicial)

```bash
go run ./cmd/argusenum -d example.com -o results.json