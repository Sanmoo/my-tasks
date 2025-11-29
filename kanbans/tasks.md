# My-Tasks GO CLI

## To Do

## 🏃 Doing

* Expiring-in parece que não funciona. Deve funcionar em list e remind.

## ✅ Done

* ~Esclarecer no help do comando `tasks list` como utilizar a option `--status`~
* ~Adicionar um filtro de `--tags` no comando `tasks list`, parecido com a option `--status` mas para possibilitar filtro por tags.~
* ~Melhorar output do comando `tasks list`, utilizando um design mais estilizado, de preferencia utilizando cores e se possível
  em um layout similar kanban (colunas para cada status). Possivelmente a largura do console será limitada para o número de
  colunas, por isso deverá existir alguma forma de scrolling horizontal.~
* ~Update the tests for the `pkg/views` package to reach 100% coverage. I can see that snapshots have not been updated
  to reflect the last improvements we did in the output.~
* ~Implement a minimum test coverage in the command `make test`. Should fail if overall coverage is below 80%.~
* ~Adicione um filtro `--expiring-in`, com alias `-ei`, que deve aceitar expressões como "1d", "2m", "5s", "10w" (respectivamente,
  1 dia, 2 minutos, 5 segundos, 10 semanas). Deve filtrar todas as tarefas com tags `@remind` ativas e tags `@due`
  com datas ou horários que serão alcançados dentro do período especificado. Os testes unitários devem ser atualizados.~
