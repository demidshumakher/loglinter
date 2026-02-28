# Log Linter

[Сборка](https://github.com/demidshumakher/loglinter/blob/master/BUILD.md)

## Дефолтный конфиг

```yaml
#.golangci.yml
version: "2"

linters:
    enable:
        - loglinter
    settings:
        custom:
            loglinter:
                type: "module"
                description: log linter.
                settings:
                    rules:
                        lowercase:
                            enabled: true
                        english_only:
                            enabled: true
                        no_special_chars:
                            enabled: true
                        sensitive_words:
                            enabled: true
                            words:
                                - password
                                - passwd
                                - secret
                                - token
                                - api_key
                                - apikey
                                - auth
                                - credential
                                - private_key
                                - access_token
                                - refresh_token
                                - bearer
                                - secret_key
                                - encryption_key
                        custom_patterns:
                            enabled: false
                            patterns:
                                - "secret info"
```

В данном примере введены все дефолтные чувствительные слова, но если их удалить и ввести другие, то будет происходить поиск только по словам из конфига, при этом если оставить список пустым, то поиск будет происходить по этим словам.

Поиск происходит в случае если идет строка + переменная, и если имя переменной есть в этом списке, то выходит предупреждение.

Кастомные паттерны ищут содержание в строке с помощью регулярного выражения

## Проект для теста
находится в `/example-project` и в `/zap-project`

## SuggestedFixes
Реализованы для заглавной буквы и специальных символов

## Анализ проекта
В проект [grafana](https://github.com/grafana/grafana) линтер не нашел проблем


В проекте [traefik](https://github.com/traefik/traefik) линтер также не нашел проблем



