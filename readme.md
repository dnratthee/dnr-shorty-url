# DNR Shorty URL

It is a simple URL shortener service that allows you to shorten long URLs.

## Technologies and Tools

- [Golang](https://golang.org/)
- [MongoDB](https://www.mongodb.com/)

## Demo

POST /

```json
{
  "uri": "https://api.sampleapis.com/codingresources/codingResources"
}
```

Response

```json
{
  "shorty": "https://s.dnratthee.me/1rqOiT0",
  "uri": "https://api.sampleapis.com/codingresources/codingResources"
}
```

GET /1rqOiT0
Response redirect to `https://api.sampleapis.com/codingresources/codingResources`
