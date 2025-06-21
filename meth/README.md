<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/meth">meth</a>
</h1>

<p align="center">Fast HTTP methods availability checker</p>

<br>
<br>

```bash
meth -u "https://target.com/paths/"
```
`OR`
```bash
cat urls.txt | meth
```

<br>
<br>

```yaml
Usage of meth:
  -c string
        Add custom methods (comma-separated)
  -fc string
        Filter out status codes (comma-separated)
  -i string
        Include only these methods (comma-separated)
  -mc string
        Show only these status codes (comma-separated)
  -s    Silent mode (only show working methods)
  -t int
        Number of threads (default 10)
  -timeout int
        Timeout in seconds (default 10)
  -u string
        Target URL
  -v    Verbose output
  -x string
        Exclude methods (comma-separated)
```
