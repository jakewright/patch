# Patch
<img align="right" width="200" src="https://storage.googleapis.com/gopherizeme.appspot.com/gophers/b32565ea974adc234d288a4dc36c994592194e35.png">

Patch is an HTTP client built on top of [net/http](https://golang.org/pkg/net/http/) that that helps you make API requests without the boilerplate.

### Features
- Automatic encoding & decoding of bodies
- Option to bring your own `http.Client{}`
- Easy asynchronous requests
- Response status code validation

## Installation

```go
go get github.com/jakewright/patch
```

## Usage

### Creating a client

The `New` function will return a client with sensible defaults

```go
c := patch.New()
```

The defaults can be overridden with Options.

```go
c := patch.New(
    // The default timeout is 30 seconds. This can be
    // changed. Setting a timeout of 0 means no timeout. 
    patch.WithTimeout(10 * time.Second),
    
    // The default status validator returns true for
    // any 2xx status code. To remove the status
    // validator, pass nil instead of a func.
    patch.WithStatusValidator(func(status int) bool {
        return status == 200
    }),
    
    // By default, request bodies are encoded as JSON.
    // This can be changed by providing a different 
    // Encoder. If a request has its own Encoder set, 
    // it will override the client's Encoder.
    patch.WithEncoder(&patch.EncoderFormURL{}),
)
```

**Custom base client**

Patch creates an `http.Client{}` which it uses to make requests. If you'd like to provide your own instance, use the `NewFromBaseClient` function.

```
bc := http.Client{}
c := NewFromBaseClient(&bc)
```

For flexibility, a custom base client doesn't have to be of type `http.Client{}`. It just has to implement the following interface. Note that the `WithTimeout` option won't work with non-standard base client types.

```go
type httpClient interface {
    Do(*http.Request) (*http.Response, error)
}
```

### Making a `GET` request

```go
user := struct{
    Name string `json:"name"`
    Age int `json:"age"
}{}

// The response is returned and also decoded into the last argument
rsp, err := client.Get(ctx, http://example.com/user/204, &user)
if err != nil {
    panic(err)
}
```

The response type embeds the original `http.Response{}` but provides some convenience functions.

```go
// Read the body as a []byte or string
b, err := rsp.BodyBytes()
s, err := rsp.BodyString()
```

The body can be read an unlimited number of times. The underlying `rsp.Body` is also available as normal.

### Making a `POST` request

The `Post()` function takes an extra argument: the body. By default, it will be encoded as JSON and an `application/json; charset=utf-8` Content-Type header will be set.

Helper functions also exist for `PUT`, `PATCH` and `DELETE`.

```go
body := struct{
    Name string `json:"name"`
    Age int `json:"age"`
}{
    Name: "Homer Simpson",
    Age: 39,
}

// If desired, the response body can be decoded into the last argument.
rsp, err := client.Post(ctx, "http://example.com/users", &body, nil)
```

_Note that the response is not decoded if the request fails, including if status code validation fails. See the section on error handling for more information._

### Making asynchronous requests
The helper functions `Get`, `Post`, `Put`, `Patch` and `Delete` are built on top the of `Send` function. You can use this directly for more control over the request, including making asynchronous requests.

```go
req := &patch.Request{
    Method: "GET"
    URL:    "http://example.com"
}

// Send is non-blocking and returns a Future
ftr := client.Send(&req)

// Do other work

// Response blocks until the response is available
rsp, err := ftr.Response()
```

### Encoding the request

By default, requests are encoded as JSON. The default encoding can be changed by using the `WithEncoder()` option when creating the client.

Encoding can also be set per-request by setting the `Encoder` field on the request struct. If this is not `nil`, it will override the client's default Encoder.

```go
req := &patch.Request{
    Encoder: &patch.EncoderFormURL{},
}
```

**JSON encoder**

The JSON encoder uses [`encoding/json`](https://golang.org/pkg/encoding/json/) to marshal the body into JSON. The Content-Type header is set to `application/json; charset=utf-8` but this can be changed by setting the `CustomContentType` field on the `EncoderJSON{}` struct.

**Form URL encoder**

The Form encoder will marshal types as follows:
1. If the body is of type `url.Values{}` or `map[string][]string`, it is encoded using [Values.Encode](https://golang.org/pkg/net/url/#Values.Encode).
2. If the body is of type `map[string]string`, it is converted to a `url.Values{}` and encoded as above.
3. If the body is of any other type, it is converted to a `url.Values{}` by [`gorilla/schema`](https://www.gorillatoolkit.org/pkg/schema) and then encoded as above.

The tag alias used by [`gorilla/schema`](https://www.gorillatoolkit.org/pkg/schema) is configurable on the `EncoderFormURL{}` struct.

The Content-Type header is set to `application/x-www-form-urlencoded` but this can be changed by setting the `CustomContentType` field on the `EncoderFormURL{}` struct.

```go
enc := &patch.EncoderFormURL{
    TagAlias: "url",
}

client, err := patch.New(patch.WithEncoder(enc))
if err != nil {
    panic(err)
}

// The body will be encoded as "name=Homer&age=39"

body := struct{
    Name string `url:"name"`
    Age int `url:"age"`
}{
    Name: "Homer Simpson",
    Age: 39,
}

rsp, err := client.Post(ctx, "http://example.com", &body, nil)
```

**Custom encoder**

A custom encoder can be provided. It must implement the following interface.

```go
type Encoder interface {
    ContentType() string
    Encode(interface{}) (io.Reader, error)
}
```

### Decoding the response

If the final argument `v` to `Get`, `Post`, `Put`, `Patch` or `Delete` is not `nil`, then the body will be decoded into the value pointed to by `v`. The decoder to use will be inferred from the response's Content-Type header. To explicitly specify a Decoder, use the convenience functions on the `Response` struct.

```go
rsp, err := client.Get(ctx, "http://example.com", nil, nil)
if err != nil {
    panic(err)
}

v := struct{...}{}

// Decode will infer the decoder from the Content-Type header.
err := rsp.Decode(&v)

// DecodeJSON will decode the body as JSON, regardless of the Content-Type header.
err := rsp.DecodeJSON(&v)

// DecodeUsing will decode the body using a custom Decoder.
err := rsp.DecodeUsing(dec, &v)
```

**Decode hooks**

Sometimes, you want to decode into different targets depending on the response status code. Arguments to the decode functions can be wrapped in a `DecodeHook` to specify for which status codes the target should be used.

```go
err := rsp.Decode(patch.On2xx(&result), patch.On4xx(&clientErr), patch.On5xx(&serverErr))
```

Decode hooks work as the final argument to the method helper functions too.

Specific status codes can be targeted using the `patch.OnStatus(404, &target)` hook. Of course, you can write your own hooks too.

### Error handling

The method helper functions `Get`, `Post`, `Put`, `Patch` and `Delete` will not try to decode the body if the `baseClient` returned an error, of if the status validator returns false.

If the request succeeds but decoding the body fails, the decoding error will be returned.

Some errors are identifiable using `errors.As()`. See `errors.go` for a list of typed errors that can be returned.

### Advanced example

Here is an example of integrating with the [GitHub API](https://developer.github.com/v3/repos/#list-repositories-for-a-user) to list repositories by user, inspired by [dghubble/sling](https://github.com/dghubble/sling).

```go
type Repository struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

type GithubError struct {
    Message string `json:"message"`
}

func (e *GithubError) Error() string {
    return e.Message
}

type RepoService struct {
    client *patch.Client
}

func NewRepoService() *RepoService {
    sv := func(status int) bool {
        if status >= 200 && status < 300 {
            return true
        }

        // Allow 4xx status codes because we
        // expect to be able to decode them
        if status >= 400 && status < 500 {
            return true
        }

        return false
    }

    return &RepoService{
        client: patch.New(
            patch.WithBaseURL("https://api.github.com"),
            patch.WithStatusValidator(sv),
        ),
    }
}

func (s *RepoService) List(ctx context.Context, username string) ([]*Repository, error) {
    path := fmt.Sprintf("/users/%s/repos", username)
    rsp, err := s.client.Get(ctx, path, nil)
    if err != nil {
        panic(err)
    }

    var repos []*Repository
    var apiErr *GithubError

    if err := rsp.DecodeJSON(On2xx(repos), On4xx(apiErr)); err != nil {
        return nil, err
    }

    return repos, apiErr
}

```

## Inspiration

Inspired by a multitude of great HTTP clients, including but not limited to:

- [dghubble/sling](https://github.com/dghubble/sling)
- [monzo/typhon](https://github.com/monzo/typhon)
- [axios/axios](https://github.com/axios/axios)

## License

[MIT License](LICENSE)
