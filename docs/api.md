# PhotoMeta API 1.0.0

This API allows you to analyze image metadata remotely.

## Base URL
The server runs by default on `http://localhost:8080`.

## Endpoints

| Method | Endpoint   | Description                                                     |
| :----- | :--------- | :-------------------------------------------------------------- |
| POST   | `/analyze` | Analyze an uploaded image file and return metadata in JSON format. |
| GET    | `/locales` | Retrieve a list of supported locales and their descriptions.    |
| GET    | `/demo`    | Serve a minimalist web page for interactive metadata analysis.  |

### POST /analyze

Upload an image file to extract its metadata.

**Request**
*   **Method**: `POST`
*   **Content-Type**: `multipart/form-data`
*   **Body**:
    *   `file`: The image file to analyze (binary).
    *   `lang`: (Optional) Language code for localized metadata (e.g., `ua`, `de`, `es`, `en`). Defaults to `en`.

**Response**
*   **Status**: `200 OK`
*   **Content-Type**: `application/json`
*   **Body**: Returns a JSON object with a flat array of localized metadata properties.

**Example Request (curl)**

```bash
curl -X POST -F "file=@/path/to/image.jpg" -F "lang=ua" http://localhost:8080/analyze
```

**Example Response**

```json
{
  "path": "",
  "name": "image.jpg",
  "format": "jpeg",
  "metadata": [
    {
      "group": "Файл",
      "raw_group": "File",
      "type": "File",
      "synonym": "Формат",
      "value": "JPEG"
    },
    {
      "group": "Файл",
      "raw_group": "File",
      "type": "File",
      "synonym": "Розмір",
      "value": "245000 bytes"
    },
    {
      "group": "Зйомка",
      "raw_group": "Shooting",
      "type": "EXIF",
      "synonym": "Витримка",
      "value": "1/60"
    }
  ]
}
```

**Error Responses**
*   `400 Bad Request`: If the file is missing or invalid.
*   `405 Method Not Allowed`: If using a method other than POST.
*   `500 Internal Server Error`: If analysis fails.

---

### GET /locales

Returns the list of available display languages.

**Request**
*   **Method**: `GET`

**Response**
*   **Status**: `200 OK`
*   **Content-Type**: `application/json`
*   **Body**: Returns a JSON array of locale objects.

**Example Request (curl)**

```bash
curl http://localhost:8080/locales
```

**Example Response**

```json
[
  {"code": "en", "description": "English"},
  {"code": "ua", "description": "Українська"},
  {"code": "ru", "description": "Русский"},
  {"code": "de", "description": "Deutsch"},
  {"code": "fr", "description": "Français"},
  {"code": "es", "description": "Español"}
]
```

**Error Responses**
*   `405 Method Not Allowed`: If using a method other than GET.
