<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/sway">sway</a>
</h1>

<p align="center">HTTP methods prober</p>

<br>
<br>

```bash
cat urls.txt | sway
```
`OR`
```bash
cat urls.txt | sway | grep '="200"'
```
`OR` <kbd>chain with other tools</kbd>
```bash
subfinder -d hilton.com -silent | httpx -silent -rl 100 -t 100 | sway
```
