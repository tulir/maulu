# mau\Lu
## Introduction
mau\Lu is a fairly simple URL shortening backend written in Go. The first versions also contain support for a frontend. The code at commit 3fc7b2d002af1da81af6a5cf2d69e8ddac415a69 is confirmed to work with a basic HTML-only frontend.

A fancy frontend can be found from [mau.lu](https://mau.lu/). I'll release the frontend properly at some point.

## API
### Requests
The API is located at /query/ and it requires a JSON payload. The payload must contain at least two fields: "action" and "url". It may optionally contain an extra field, "short-request", to have custom short URLs. The extra field will be ignored when unshortening.

Currently supported actions:
 * `shorten` - Shorten the requested URL.
 * `google` - Create a [LMGTFY](http://lmgtfy.com/) URL using the data in the URL field as the search query and shorten the created URL.
 * `duckduckgo` - The same as `google`, but with [LMDDGTFY](https://lmddgtfy.net/), a tool similiar to LMGTFY that uses [DuckDuckGo](https://duckduckgo.com/)
 * `unshorten` - Unshortens the given mau\Lu URL.

Example shortening request:
```json
{
    "action": "shorten",
    "url": "https://git.maunium.net/Tulir293/maulu",
    "short-request": "maulu-git"
}
```

### Responses
The API will respond with a JSON that contains either the fields `error` and `error-long` or just `result`. The `error-long` field is a human-readable version, whilst the `error` field is a simple error tag.

When successful, `unshorten` will return the long URL and `shorten`, `google` and `duckduckgo` will return the short key that must be appended to the mau\Lu URL.

Possible errors:
 * `action` - The given action couldn't be identified.
 * For action `unshorten`
  * `notshortened` - The given URL is not a short URL.
  * `toolong` - The given URL is too long to be a short URL.
  * `notfound` - The given short URL doesn't exist.
 * For actions `shorten`, `google` and `duckduckgo`
  * `illegalchars` - The requested short key (`short-request`) contains illegal characters.
  * `alreadyshortened` - The given URL is an URL that has already been shortened using mau\Lu.
  * `invalidprotocol` - The URL uses an unidentified protocol.
  * `toolong` - The given URL is too long.
  * `alreadyinuse` - The requested short key (`short-request`) is already in use.
