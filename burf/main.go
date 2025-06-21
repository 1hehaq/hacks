package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var exts = []string{
	".bak", ".old", ".backup", ".zip", ".tar.gz", ".sql", ".db", ".env", ".config", ".log", ".gz", ".rar", ".7z", ".tgz", ".sqlite", ".key", ".crt", ".jar", ".war", ".ini", ".cfg", ".yml", ".json", ".xml", ".txt", ".csv", ".dat", ".tmp", ".temp", ".orig", ".save", ".swp", ".lock", ".pid", ".cache", ".sess", ".conf", ".properties", ".settings", ".prefs", ".data", ".dump", ".backup.zip", ".backup.tar", ".backup.gz", ".pem", ".p12", ".pfx", ".jks", ".keystore", ".cer", ".der", ".csr", ".ovpn", ".ppk", ".pub", ".rsa", ".dsa", ".id_rsa", ".id_dsa", ".ssh", ".passwd", ".shadow", ".htpasswd", ".secret", ".secrets", ".credentials", ".creds", ".password", ".passwords", ".pwd", ".token", ".tokens", ".auth", ".oauth", ".jwt", ".session", ".sessionid", ".cookie", ".cookies", ".apikey", ".api_key", ".access_token", ".refresh_token", ".private", ".confidential", ".sensitive", ".internal", ".admin", ".root", ".debug", ".trace", ".error", ".exception", ".crash", ".core", ".mem", ".heap", ".stack",
}

func main() {
	word := flag.String("s", "", "string")
	flag.Parse()

	var words []string
	if *word != "" {
		words = append(words, *word)
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			if line := strings.TrimSpace(scanner.Text()); line != "" {
				words = append(words, line)
			}
		}
	}

	for _, w := range words {
		output(w)
	}
}

func output(word string) {
	for _, ext := range exts {
		fmt.Printf("%s%s\n", word, ext)
	}
}
