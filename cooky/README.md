<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/cooky">cooky</a>
</h1>

<p align="center">cookie decoder and encoder detector</p>

<br>
<br>

```bash
cooky -u example.com
```
`OR`
```bash
echo "example.com" | cooky
```

<br>
<br>

```yaml
usage: cooky [-u url] [-t timeout] [-c concurrency] [-json]
  -u     target url
  -t     timeout (default 10s)
  -c     concurrency (default 10)
  -json  output raw JSON format
reads from stdin if no flags given
```
