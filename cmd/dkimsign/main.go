package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/driusan/dkim"

	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func signmessage(sig dkim.Signature, key *rsa.PrivateKey, unix bool, dotstuffed bool, hdronly bool) error {
	r := dkim.NormalizeReader(os.Stdin)
	if dotstuffed {
		r.Unstuff()
	}
	file, err := dkim.FileBuffer(r)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())

	var nl string
	if unix {
		nl = "\n"
	}
	if hdronly {
		return dkim.SignedHeader(sig, file, os.Stdout, key, nl)
	}
	return dkim.SignMessage(sig, file, os.Stdout, key, nl)
}

func main() {
	var canon string = "relaxed/relaxed"
	var s, domain string
	var headers string
	var unstuff bool
	var headeronly bool
	flag.StringVar(&canon, "c", "relaxed/relaxed", "Canonicalization scheme")
	flag.StringVar(&s, "s", "", "Domain selector")
	flag.StringVar(&domain, "d", "", "Domain name")
	flag.StringVar(&headers, "h", "From:Subject:To:Date", "Colon separated list of headers to sign")
	flag.BoolVar(&unstuff, "u", false, "Assume input is already SMTP dot stuffed when calculating signature and un dot-stuff it while printing")
	nl := flag.Bool("n", false, `Print final message with \n instead of \r\n line endings`)
	flag.BoolVar(&headeronly, "hd", false, "Only print the header, not the whole message after signing")
	privatekey := flag.String("key", "", "Location of PEM encoded private key")
	flag.Parse()

	if domain == "" || s == "" {
		fmt.Fprintln(os.Stderr, "Selector and domain are required")
		os.Exit(1)
	}
	var key *rsa.PrivateKey
	kf, err := os.Open(*privatekey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open private key: %v\n", err)
		os.Exit(1)

	}
	defer kf.Close()
	keyfile, err := ioutil.ReadAll(kf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could read private key: %v\n", err)
		os.Exit(1)
	}

	pemblock, _ := pem.Decode(keyfile)
	if pemblock == nil || pemblock.Type != "RSA PRIVATE KEY" {
		fmt.Fprintln(os.Stderr, "Could read private key or unsupported format")
		os.Exit(1)
	}
	key, err = x509.ParsePKCS1PrivateKey(pemblock.Bytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not parse private key: %v\n", err)
		os.Exit(1)
	}

	sig, err := dkim.NewSignature(canon, s, domain, strings.Split(headers, ":"))

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := signmessage(sig, key, *nl, unstuff, headeronly); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
