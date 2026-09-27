Status: ready-for-agent

# Spec: `mt defer` aceita prazos relativos em anos e meses

## Problem Statement

O `mt defer <id> <quando>` aceita o absoluto `YY-MM-DD HH:MM` e relativos
`+<n><unit>` com `d` (dias), `w` (semanas) e `h` (horas) — todos durações
exatas. Não há como adiar por um mês ou ano sem contar dias na mão, e o
absoluto só alcança 2000–2099 (o ano é de dois dígitos). `mt defer <id> +6y`
falha com "invalid defer duration".

## Solution

O `Parse` passa a aceitar duas unidades de calendário: `m` (meses) e `y`
(anos). Diferente de `d`/`w`/`h`, que somam uma duração exata, `m`/`y` somam
meses/anos de calendário: o dia do mês alvo é preservado e limitado ao
último dia quando o mês é mais curto (31/01 +1m → 28/02; 29/02 +1y →
28/02), e a hora é preservada. O alvo precisa caber no ano de quatro dígitos
que o `NaiveLayout` guarda (0000–9999); um alvo além disso é rejeitado como
duração inválida, sem overflow. Nada mais muda: o campo gravado, o exit
code e o pipeline de mutação ficam como estão.

## User Stories

1. As a usuário, quero `mt defer <id> +6y` para adiar por anos, sem calcular
   a data absoluta (que só alcança 2099).
2. As a usuário, quero `mt defer <id> +6m` para adiar por meses.
3. As a usuário, quero que adiar por um mês caia no mesmo dia do mês
   seguinte quando ele existe, e no último dia quando não existe (31/01
   +1m → 28/02), para o alvo nunca "pular" um mês.
4. As a usuário, quero que adiar por um ano a partir de 29/02 caia em
   28/02, para a data ser sempre representável.
5. As a usuário, quero continuar adiando com `+2d`, `+1w`, `+3h` exatamente
   como hoje (duração exata, sem calendário).
6. As a usuário, quero que um prazo que ultrapasse 9999 seja rejeitado com
   erro de duração, e não grave um `deferred_until` que o resto do CLI não
   consegue parsear.
7. As a usuário, quero que o help do comando e o README listem as novas
   unidades.

## Implementation Decisions

- `internal/deferral.Parse` reconhece `m` e `y` no relativo `+<n><unit>`.
  Conflito de `m` com "minutos" não existe: minutos não são unidade do
  comando (a menor é `h`), então `m` é livre para "mês".
- Calendário não é duração: `m`/`y` usam aritmética de calendário
  (`AddDate`-equivalente) com clamp de dia para o último dia do mês alvo,
  em vez de somar um número fixo de horas.
- O alvo é validado contra o ano máximo representável (`maxYear = 9999`): a
  checagem acontece antes da soma para que contagens próximas do limite do
  `Atoi` não estourem o `int`. Alvo fora da faixa → `invalidDuration`, o
  mesmo erro das outras durações malformadas.
- A checagem de overflow das durações exatas (`maxDuration`) fica como está.
- Mensagens de erro/uso passam a listar `m` (months) e `y` (years).
- Documentação: `deferLong`, a mensagem de args, o README e o
  `defer.feature` citam as novas unidades. O spec do `mt-tracker` (histórico
  do ticket 07) fica como está.

## Testing Decisions

Seams: o único seam novo é o já estabelecido.

- **Seam 2 — `internal/deferral.Parse` (primário)**: teste unitário
  black-box. Casos novos: `+1y`/`+6y`, `+1m`/`+6m`/`+12m`/`+18m`, virada de
  ano (`+5m` de agosto → janeiro), clamp de fim de mês (31/01 +1m; 31/01
  +1m em ano bissexto → 29/02; 29/02 +1y → 28/02), limites exatos do ano
  (`+7973y` válido / `+7974y` erro; `+95680m` válido / `+95681m` erro) e
  contagens que estouram o `Atoi` para `m`/`y`. O hint de erro passa a
  cobrar a citação de `m`/`y`.
- **Seam 1 — processo CLI (regressão)**: o `defer.feature` ganha um cenário
  que adia com `+6y` e `+6m` e checa o formato do `deferred_until` gravado;
  os cenários existentes de `+2d`/absoluto são a regressão das durações
  exatas.

## Out of Scope

- Suporte a minutos (`+30min`) ou a qualquer outra unidade.
- Ajustar o `absoluteLayout` de dois dígitos (o `+1y` é a via para além de
  2099).
- Mudar `total-max`/expiração (`IsFutureDeferred`), `undefer` ou `overdue`.
- Reescrever o spec/ticket 07 do `mt-tracker`.

## Further Notes

- A unidade de mês é `m` (não `mo`): o CLI não tem minutos, então não há
  ambiguidade, e mantém a simetria de uma letra (`d`/`w`/`h`/`m`/`y`).
- O clamp de fim de mês é decisão local: `time.AddDate` do Go normaliza
  31/01 +1m para 02/03 (ou 03/03), pulando fevereiro — alvo errado para
  uma Deferral.
