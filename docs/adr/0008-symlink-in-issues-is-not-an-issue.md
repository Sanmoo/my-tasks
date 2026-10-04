# Um symlink em issues/ não é uma Issue

O acesso aos arquivos de Issue vivia espalhado pelo wiring do CLI, e cada
comando decidia por conta própria como tratar o mesmo estado em disco. Um
symlink `issues/*.md` recebia **seis respostas diferentes**: a listagem em
lote (`list`, `check`, `ready`, `overdue`, `pick-next`, `place`, `undefer`
em lote, `prioritize`) derrubava o Vault inteiro com "is a symbolic link";
a completion de shell **oferecia** o symlink como ID e a alocação o tratava
como nome ocupado; leitura e escrita pontuais o recusavam; `mt show` o
**seguia**; `mt edit` o seguia e entregava o alvo ao `$EDITOR`; `mt create`
podia **escrever através** dele. O risco era concreto: um link plantado em
`issues/` permitia que o mt lesse, editasse e sobrescrevesse um arquivo
fora do Vault — enquanto a listagem recusava o Vault inteiro.

O módulo `internal/issuefiles` passa a ser o único dono do acesso a esses
arquivos, e a política passa a ser uma só: **um symlink dentro de `issues/`
não é uma Issue**. Ele é **pulado** por toda listagem — `List`, `ListFiles`
e `IDs`, esta última servindo a completion —, então um link perdido não
derruba mais o Vault e o TAB não oferece um ID que os comandos recusam. E
é **recusado** por `Read`, `Mutate` e `Edit` com a mesma razão ("issue <id>
is not a regular file"), antes de qualquer leitura ou escrita: `Edit`
valida antes de o caminho sair do módulo, e `Mutate` relê o arquivo como
arquivo regular antes de gravar. `Create` faz o próprio scan e trata todo
`*.md` que não é diretório — symlink incluso — como nome ocupado, de modo
que um link nunca é reutilizado como ID de uma Issue nova; o `open` ainda
usa `O_CREATE|O_EXCL` mais `O_NOFOLLOW`, fechando a corrida entre descobrir
e abrir. No Unix, `O_NOFOLLOW` no `open` fecha a mesma corrida na leitura e
na escrita.

O resto da política de arquivo é preservado byte a byte e agora também
está escrito: uma entrada `.md` que é **diretório** continua sendo pulada
(um diretório acidental não vira Issue fantasma); uma entrada
**não-regular que não é diretório** (FIFO, device) continua **falhando
alto** (um Vault quebrado não passa silencioso); um `issues/` **ilegível**
continua produzindo erro claro; e `Open` **não carrega `mt.yaml`** — o
módulo recebe o diretório e não decide o que é um Vault, então os comandos
que hoje operam sem o arquivo de configuração continuam operando sem ele.
Num Vault sem symlinks, nada muda: a mudança de comportamento é
exatamente esta, isolada num commit.
