<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/burf">burf</a>
</h1>

<p align="center">combine words with sensitive file extensions</p>

<br>
<br>

```bash
echo -e "github\ngithub.com" | burf
```
```bash
echo "api" | burf | sed 's/^/\/v1\//'
```

<br>

<kbd>chain with ffuf</kbd>
```bash
burf -s github.com | ffuf -w - -u https://github.com/FUZZ
```

<br>
<br>

```yaml
Usage of burf:
  -s string
        string
```
