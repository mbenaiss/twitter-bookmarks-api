# Twitter Bookmarks API Documentation

## Authentication
All endpoints require authentication via the `Authorization` header with a Bearer token.

Example:
```
Authorization: Bearer your_access_token
```

## Endpoints

### Get User Bookmarks
Retrieves all bookmarks for a specific user.

```
GET /bookmarks/:user_id
```

#### Parameters
- `user_id` (path parameter): The Twitter user ID to fetch bookmarks for

#### Response
```json
{
  "bookmarks": [
    {
      "id": "string",
      "tweet_id": "string",
      "text": "string",
      "created_at": "2024-12-05T12:00:00Z",
      "author": {
        "id": "string",
        "username": "string",
        "name": "string"
      }
    }
  ],
  "next_token": "string"
}
```

### Get Filtered Bookmarks
Retrieves bookmarks for a user after a specific date.

```
GET /bookmarks/:user_id/filter
```

#### Parameters
- `user_id` (path parameter): The Twitter user ID to fetch bookmarks for
- `after` (query parameter): RFC3339 formatted date to filter bookmarks after

Example:
```
GET /bookmarks/123456789/filter?after=2024-12-01T00:00:00Z
```

#### Response
Same as Get User Bookmarks endpoint.

## Error Responses

### 401 Unauthorized
```json
{
  "error": "Authorization header is required"
}
```

### 400 Bad Request
```json
{
  "error": "Invalid date format. Use RFC3339"
}
```

### 500 Internal Server Error
```json
{
  "error": "Failed to fetch bookmarks",
  "details": "error description"
}
```
