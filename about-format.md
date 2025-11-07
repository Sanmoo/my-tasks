# About the Markdown Format for this task manager

- The file `kanbans/domestic_alt.md` describes an example of file format that this application is supposed to support.
- Lines starting with "#" (header 1) declare a new "Project".
- Under projects, lines starting with "##" (header 2) declare Phases for tasks inside this project. They are like columns in a Kanban board.
- Under each phase, each high level bullet point declares a task
- Under each task, each sub bullet point that is not initiated with "@" can be free text and should be interpreted as a comment for the task.
- Under each task, each sub bullet point that starts with "@remind" must be followed by a date or time format between parenthesis, like "(25-10-23 25:00:00)"
  or "(25-10-23)", and represents a reminder or time for which a notification should be sent regarding the related task.
- Under each task, each sub bullet point that starts with "@reminded" should be interpreted like the ones started with "@remind", but the "ed" at the end
  indicates that the user already acknowledged that notification and there is no need to the notification to be sent.
- Under each task, each sub bullet point that starts with "@tags" should be interpreted like a space separated list of words that represent tags for the related
task. There should be at maximum one "@tag" directive for each task.
- Under each task, each sub bullet point that starts with "@due" must be followed by a date or time format between parenthesis, like "(25-10-23 25:00:00)" or
  "(25-10-23)", and represents the time or date before which the related task should be completed. There should be at maximum one @due directive per task.
