<h1 align="center">
  <a href="https://github.com/1hehaq/hacks/tree/main/rotl">rotl</a>
</h1>

<p align="center">rotate and transform urls for bypass</p>

<br>
<br>

```bash
echo "https://github.com/api/v1" | rotl
```
`OR`
```bash
echo "https://github.com/admin" | rotl | httpx -mc 200,403,301
```
