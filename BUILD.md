# Сборка
Скопировать репозиторий

```bash
git clone https://github.com/demidshumakher/loglinter.git
```

```yml
# .custom-gcl.yml
version: v2.10.1
plugins:
    - module: "github.com/demidshumakher/loglinter"
      path: /path/to/linter
```
указать путь до папки `/src`

```bash
golangci-lint custom --name golangci-lint-loglinter --destination .
```

После чего появится бинарный файл `golangci-lint-loglinter`

Для работы с линтером в конфиге надо указать:
```yml
# .golangci.yml
version: "2"

linters:
    enable:
        - loglinter
    settings:
        custom:
            loglinter:
                type: "module"
                description: log linter.
```

Для запуска необходимо прописать в директории проекта

```bash
golangci-lint-loglinter run -c
```

С указанием конфига
```bash
golangci-lint-loglinter run -c "../.golangci.yml"   
```
