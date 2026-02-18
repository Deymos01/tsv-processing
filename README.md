# TSV Processor

Go-сервис, который отслеживает директорию на наличие `.tsv` файлов, парсит их,
сохраняет данные в PostgreSQL и генерирует RTF-отчёты для каждого `unit_guid`.

## Требования

- Go 1.25+
- Docker & Docker Compose

## Быстрый старт
```bash
# 1. Склонируйте репозиторий
git clone https://github.com/Deymos01/tsv-processing.git
cd tsv-processing

# 2. Скопируйте и отредактируйте файл конфигурации
cp config.yaml.example config.yaml

# 3. Запустите PostgreSQL и сервис с помощью docker
docker compose up --build
```


## HTTP API

### GET /api/v1/messages

Возвращает список сообщений с пагинацией, отфильтрованных по `unit_guid`.

**Параметры запроса:**

| Parameter   | Required | Default | Description                                 |
|-------------|----------|---------|---------------------------------------------|
| `unit_guid` | yes      | —       | GUID устройства для фильтрации              |
| `page`      | no       | `1`     | Номер страницы                              |
| `limit`     | no       | `20`    | Количество на одной странице (макс. `1000`) |

**Пример запроса:**
```bash
curl "http://localhost:8080/api/v1/messages?unit_guid=abc-123&page=1&limit=20"
```

**Пример ответа:**
```json
{
  "data": [
    {
      "id": 1,
      "number": 1,
      "mqtt": "some/topic",
      "unit_guid": "abc-123",
      "message_text": "Example message",
      "message_context": "production",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 1,
    "total_pages": 1
  }
}
```

## Формат TSV-файла

Поместите `.tsv` файлы в директорию, указанную в `input_dir`. Файл должен содержать
строку с комментарием, строку заголовка, а затем строки с данными со следующими
колонками (разделитель — табуляция):

| # | Column                  |
|---|-------------------------|
| 0 | number                  |
| 1 | mqtt                    |
| 2 | inv_id                  |
| 3 | unit_guid               |
| 4 | message_id              |
| 5 | message_text            |
| 6 | context                 |
| 7 | message_class           |
| 8 | message_level           |
| 9 | variable_zone           |
|10 | variable_address        |
|11 | use_as_block_start      |
|12 | type                    |
|13 | bit_number_in_register  |
|14 | invert_bit              |

## Выходные файлы

После успешной обработки RTF-файлы сохраняются в директорию `output_dir`:

- `<unit_guid>.rtf` — отчёт со всеми сообщениями для данного устройства
- `<source_filename>_error.rtf` — отчёт об ошибке, если парсинг завершился неудачно