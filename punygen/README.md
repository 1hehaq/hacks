<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/punygen">punygen</a>
</h1>

<p align="center">homoglyphs and punycode generator</p>

<br>
<br>

```bash
punygen -i letter/word
```
`OR`
```bash
echo "word/letter" | punygen
```

<br>
<br>

```yaml
usage: punygen [-i input] [-v] [-m max]
  -i  input (letter or word)
  -v  verbose (show punycode)
  -m  max combinations (default 1000)
reads from stdin if no flags given
```
