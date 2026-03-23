# Relatorio de Carga (k6)

Neste relatorio, eu apresento os testes de carga que executei e a leitura que fiz dos resultados.

## Configuracao dos testes
- Script: `loadtest/telemetria.js`
- Endpoint: `POST /telemetria`
- Campos enviados: `id_dispositivo`, `instante`, `tipo_sensor`, `natureza_leitura`, `valor_coletado`

## Rodada base
Comando utilizado:
```bash
k6 run loadtest/telemetria.js
```

Resultado base (com `sleep`):
- Cenario A (`0.50` CPU por servico): `216.80 req/s`, p95 `6.87 ms`, erro `0.00%`
- Cenario B (`0.25` CPU por servico): `217.64 req/s`, p95 `6.35 ms`, erro `0.00%`

Obs: aqui quase nao muda porque o proprio script freia as requisicoes (`sleep(0.1)`).

## Rodada de estresse (sem sleep)
Configuracao usada:
- `SLEEP_SECONDS=0`
- `MAX_VUS=120`
- `RAMP_SECONDS=20`
- `HOLD_SECONDS=40`
- `RAMP_DOWN_SECONDS=20`

Resultados:
| Cenario | CPU backend+consumer | Vazao (`http_reqs`) | p95 (`http_req_duration`) | Erro (`http_req_failed`) | Reqs aceitas |
|---|---:|---:|---:|---:|---:|
| A | 0.25 | 1054.74 req/s | 196.89 ms | 0.00% | 84370 |
| B | 0.50 | 1689.63 req/s | 106.49 ms | 0.00% | 135160 |

Leitura rapida:
- ganho de vazao de `0.25` pra `0.50`: ~`+60.2%`
- latencia p95 caiu de `196.89 ms` pra `106.49 ms`

## Evidencia de fila e banco
- Cenario A: backlog apos teste `47241` nao-confirmadas, drenou pra `0` em ~50s.
- Cenario B: backlog apos teste `106997` nao-confirmadas, drenou pra `0` em ~80s.
- Eu conferi que a persistencia no banco bateu com as requisicoes aceitas:
  - A: `+84370` linhas
  - B: `+135160` linhas

## Estimativa por regra de 3 (so aproximacao)
Formula:
`estimativa_1_core = vazao_medida * (1.0 / cpu_usada)`

Exemplo:
- com `0.25` CPU e `1054.74 req/s`, estimativa simples em `1 core` seria `4218.96 req/s`

Importante: eu tratei isso como estimativa inicial; nao substitui medicao real.

## Gargalo e melhorias
No estresse, eu observei backlog mostrando que consumidor/banco viram limitantes antes da API cair.

Melhorias que eu faria:
1. subir mais replicas do consumidor
2. criar DLQ + politica de retry
3. ajustar indice/insert em lote no PostgreSQL
