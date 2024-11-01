package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/indece-official/go-ebcdic"
)

var (
	flagCodepage = flag.String("c", "037", "EBCDIC-Codepage to use (supported: 037 / 273 / 500 / 1140 / 1141 / 1148)")
	flagEncode   = flag.Bool("e", false, "Encode input instead of decoding it")
)

func main() {

	flag.Parse()

	codePage, err := codePage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing codepage: %s\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {

		if *flagEncode {
			out, err := ebcdic.Encode(scanner.Text(), codePage)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding data: %s\n", err)
				continue
			}
			fmt.Fprintln(os.Stdout, string(out))
			continue
		}
		out, err := ebcdic.Decode(scanner.Bytes(), codePage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding data: %s\n", err)
			continue
		}
		fmt.Fprintln(os.Stdout, out)
	}
}

func codePage() (int, error) {

	codepage := 0
	switch strings.ToLower(*flagCodepage) {
	case "37", "037", "ebcdic037":
		codepage = ebcdic.EBCDIC037
	case "273", "ebcdic273":
		codepage = ebcdic.EBCDIC273
	case "500", "ebcdic500":
		codepage = ebcdic.EBCDIC500
	case "1140", "ebcdic1140":
		codepage = ebcdic.EBCDIC1140
	case "1141", "ebcdic1141":
		codepage = ebcdic.EBCDIC1141
	case "1148", "ebcdic1148":
		codepage = ebcdic.EBCDIC1148
	default:
		return 0, fmt.Errorf("Unsupported codepage %s", *flagCodepage)
	}
	return codepage, nil
}
