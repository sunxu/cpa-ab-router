package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/sunxu/cpa-ab-router/internal/classifier"
)

func main() {
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	result := classifier.ClassifyJSON(body)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(result)
}
