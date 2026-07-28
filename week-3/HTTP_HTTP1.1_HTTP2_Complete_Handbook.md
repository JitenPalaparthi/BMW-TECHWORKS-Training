# HTTP Protocol Handbook (HTTP/1.1 and HTTP/2)

# Table of Contents

1.  Introduction
2.  Client-Server Model
3.  URI, URL, URN
4.  HTTP Message Format
5.  Request Line
6.  Status Line
7.  HTTP Methods
8.  Idempotency & Safety
9.  Request Headers
10. Response Headers
11. Entity Headers
12. Content Negotiation
13. MIME Types
14. Transfer-Encoding
15. Content-Length
16. Chunked Transfer Encoding
17. Compression
18. Caching
19. Cookies
20. Sessions
21. Authentication
22. Authorization
23. Conditional Requests
24. Range Requests
25. Connection Management
26. Keep-Alive
27. Persistent Connections
28. Pipelining
29. Proxies
30. Reverse Proxies
31. Gateways
32. Load Balancers
33. HTTPS & TLS
34. HTTP/1.1 Architecture
35. HTTP/2 Architecture
36. HPACK Header Compression
37. Multiplexing
38. Flow Control
39. Server Push (Historical)
40. Prioritization
41. Common Status Codes
42. REST Best Practices
43. Debugging HTTP
44. Wireshark & curl Examples
45. Interview Questions

------------------------------------------------------------------------

# 1. Introduction

HTTP (HyperText Transfer Protocol) is a stateless application-layer
protocol used to transfer representations of resources between clients
and servers.

Protocol stack:

Application → HTTP

Transport → TCP (HTTP/1.1, HTTP/2)

Transport → QUIC (HTTP/3)

Network → IP

Link → Ethernet / WiFi

------------------------------------------------------------------------

# Stateless

Each request is independent.

Servers generally do not remember previous requests unless cookies,
tokens, or sessions are used.

------------------------------------------------------------------------

# Request Structure

    GET /products/10 HTTP/1.1
    Host: example.com
    User-Agent: curl/8.0
    Accept: application/json

Body is optional.

------------------------------------------------------------------------

# Response Structure

    HTTP/1.1 200 OK
    Content-Type: application/json
    Content-Length: 52

    {
    "id":10
    }

------------------------------------------------------------------------

# HTTP Methods

GET POST PUT PATCH DELETE HEAD OPTIONS TRACE CONNECT

Safe: GET HEAD OPTIONS

Idempotent: GET PUT DELETE HEAD OPTIONS

------------------------------------------------------------------------

# Common Request Headers

Host User-Agent Accept Accept-Encoding Accept-Language Authorization
Cookie Origin Referer Content-Type Content-Length If-None-Match
If-Modified-Since Range Connection

Purpose and examples:

Host identifies the virtual host.

Accept tells the server acceptable media types.

Accept-Encoding advertises supported compression:

    Accept-Encoding: gzip, br, deflate

------------------------------------------------------------------------

# Common Response Headers

Content-Type

Content-Length

Server

Date

Location

Cache-Control

Expires

ETag

Last-Modified

Set-Cookie

Content-Encoding

Transfer-Encoding

WWW-Authenticate

Vary

Retry-After

------------------------------------------------------------------------

# MIME Types

text/html

application/json

application/xml

text/plain

image/png

image/jpeg

application/pdf

multipart/form-data

application/octet-stream

------------------------------------------------------------------------

# Compression

Goal: Reduce bandwidth.

Algorithms:

gzip

deflate

Brotli (br)

zstd (modern deployments)

Request:

    Accept-Encoding: gzip, br

Response:

    Content-Encoding: gzip

Compression happens after the body is generated and before transmission.

Benefits:

• Lower latency

• Less bandwidth

• Faster downloads

Avoid compressing JPEG, MP4, ZIP, PNG because they are already
compressed.

------------------------------------------------------------------------

# Chunked Transfer Encoding

When size is unknown.

    Transfer-Encoding: chunked

Example:

    7
    Mozilla
    9
    Developer
    0

------------------------------------------------------------------------

# Caching

Headers:

Cache-Control

ETag

Expires

Last-Modified

If-None-Match

If-Modified-Since

304 Not Modified saves bandwidth.

------------------------------------------------------------------------

# Cookies

    Set-Cookie:
    SESSION=abc123;
    HttpOnly;
    Secure;
    SameSite=Lax

Flags:

Secure

HttpOnly

SameSite

Expires

Max-Age

------------------------------------------------------------------------

# Authentication

Basic

Bearer JWT

Digest

Mutual TLS

OAuth2

OpenID Connect

------------------------------------------------------------------------

# HTTPS

HTTP over TLS.

Provides:

Encryption

Integrity

Authentication

Handshake:

ClientHello

ServerHello

Certificate

Key Exchange

Encrypted HTTP

------------------------------------------------------------------------

# HTTP/1.1

Major improvements:

Persistent connections

Host header

Chunked transfer

Pipelining (rare)

Cache improvements

Virtual hosting

Limitations:

Head-of-line blocking

One request per TCP connection at a time (effectively, without
multiplexing)

Large repeated headers

Many TCP connections

------------------------------------------------------------------------

# HTTP/2

Binary protocol instead of text.

Features:

Multiplexing

Streams

Frames

Header Compression (HPACK)

Flow Control

Prioritization

One TCP connection carries many streams simultaneously.

Frame types include:

DATA

HEADERS

SETTINGS

WINDOW_UPDATE

PING

RST_STREAM

GOAWAY

------------------------------------------------------------------------

# HPACK

Compresses repeated headers.

Example:

    Authorization
    Content-Type
    Accept

sent once, indexed later.

Reduces bandwidth significantly.

------------------------------------------------------------------------

# Multiplexing

HTTP/1.1

Request1 ----\>

(wait)

Response1

Request2

HTTP/2

Stream1

Stream2

Stream3

All simultaneously over one TCP connection.

------------------------------------------------------------------------

# Flow Control

Receiver advertises window size.

Sender cannot exceed available window.

WINDOW_UPDATE increases the window.

------------------------------------------------------------------------

# Server Push

Allowed server to proactively send assets.

Now deprecated by browsers because of poor practical benefit.

------------------------------------------------------------------------

# Useful curl Commands

GET

``` bash
curl http://localhost:8080/api/products
```

POST

``` bash
curl -X POST http://localhost:8080/api/products \
-H "Content-Type: application/json" \
-d '{"name":"Laptop"}'
```

Show headers

``` bash
curl -I http://localhost
```

Compressed

``` bash
curl --compressed http://localhost
```

HTTP/2

``` bash
curl --http2 https://example.com
```

------------------------------------------------------------------------

# Debugging Tools

curl

Postman

Chrome DevTools

Wireshark

tcpdump

mitmproxy

------------------------------------------------------------------------

# Interview Topics

Difference between HTTP and HTTPS

Persistent Connections

Chunked Encoding

Compression

Caching

ETag

Cookies vs Sessions

JWT

HTTP/1.1 vs HTTP/2

HPACK

Multiplexing

Flow Control

Range Requests

Conditional Requests

Connection Keep-Alive

Status Codes

REST best practices
