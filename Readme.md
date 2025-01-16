Rapid Api
==========================

## Description

Rapid Api tool is a simple API backend that allowed to quickly and easily to create/use CRUD Rest APIs.

## Quick Start

### Intro

Rapid Api tool uses Dynamodb as the database storage.
You need to create DynamoDb table on aws and also create IAM user with access to the table.

### Install

```bash
go install github.com/apibrew/rapid-api/cmd/rapid-api@latest
```

### Setup

You need to copy config.example.json to config.json and configure it.

You need to configure dynamodb and IAM details in config.json

### Run

```bash
rapid-api -config config.json
```

## Examples

### Create a book

```http request
POST /books
Content-Type: application/json

{
    "title": "The Lord of the Rings",
    "author": "J.R.R. Tolkien",
    "year": 1954
}
```

Result:

```json
{
  "author": "J.R.R. Tolkien",
  "path": "/books/1",
  "title": "The Lord of the Rings",
  "year": 1954
}
```

### Get a book

```http request
GET /books/1
```

Result:

```json
{
  "author": "J.R.R. Tolkien",
  "path": "/books/1",
  "title": "The Lord of the Rings",
  "year": 1954
}
```

### Get all books

```http request
GET /books
```

Result:

```json
[
  {
    "author": "J.R.R. Tolkien",
    "path": "/books/1",
    "title": "The Lord of the Rings",
    "year": 1954
  }
]
```

### Update a book

```http request
POST /books/1
Content-Type: application/json

{
    "title": "The Lord of the Rings",
    "author": "J.R.R. Tolkien",
    "year": 1955
}
```

Result:

```json
{
  "author": "J.R.R. Tolkien",
  "path": "/books/1",
  "title": "The Lord of the Rings",
  "year": 1955
}
```

### Delete a book

```http request
DELETE /books/1
```

### Create sub record on book (e.g. comments)

```http request
POST /books/1/comments
Content-Type: application/json

{
    "comment": "Great book!"
}
```

Result:

```json
{
  "comment": "Great book!",
  "path": "/books/1/comments/1"
}
```

### Get book with sub records

```http request
GET /books/1
```

Result:

```json
{
  "author": "J.R.R. Tolkien",
  "comments": [
    {
      "comment": "Great book!",
      "path": "/books/1/comments/1"
    }
  ],
  "path": "/books/1",
  "title": "The Lord of the Rings",
  "year": 1955
}
```

The magic happens here, it automatically handles all the relationships between the apis.