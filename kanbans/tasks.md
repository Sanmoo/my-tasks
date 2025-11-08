# My-Tasks GO CLI

## To Do

## 🏃 Doing

* Creation of `remiders` command and activate it by default
  * @remind (20-01-01 01:00) Check something

## ✅ Done

* Support `yaml` configuration file (loaded in `~/.config/tasks/tasks.conf.yaml`)
  * Should support properties like:

    ```yaml
    project:
      aliases:
        main: "Super Cool Project Name"
        pfw: "Project From Work!"
      files:
        - "$HOME/kanbans/main.md"
        - "/home/fulano/kanbans/pfw.md"
    ```

    `project.aliases` define a relation of short names that can be used to refer to project names when calling the app.
    `project.files` define a list of paths to files that the app can use to find the data
  * Testing
* Support the command `tasks <project_name> list`
  * Should list all tasks from `<project_name>`, segregated by Phase / section.
