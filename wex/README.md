<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/wex">wex</a>
</h1>

<p align="center">keyword extractor for wordlist</p>

<br>
<br>

<kbd>-url</kbd>
```bash
cat urls.txt | wex -url
```
<kbd>-js</kbd>
```bash
cat urls.txt | grep '\.js$' | wex -js
```

<br>
<br>

```yaml
Usage of wex:
  -js
        extract js keywords (supported only for .js files)
  -url
        extract path/parameter from urls
```
