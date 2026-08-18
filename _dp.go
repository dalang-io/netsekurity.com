package main

import (
	"fmt"
	"html/template"
	"os"
	"strings"
)

// Usage: go run dashparse.go <file.go>
func main() {
	file := "/tmp/nsk_build/dashboard.go"
	if len(os.Args) > 1 {
		file = os.Args[1]
	}
	b, err := os.ReadFile(file)
	if err != nil { fmt.Println("READ ERR", err); os.Exit(1) }
	s := string(b)
	i := strings.Index(s, "const dashboardHTML")
	if i < 0 { fmt.Println("no const"); os.Exit(1) }
	j := strings.Index(s[i:], "`{{define") + i
	endRel := strings.Index(s[j+1:], "`")
	src := s[j+1 : j+1+endRel]
	t, err := template.New("dashboard").Funcs(template.FuncMap{"cssHash": func() string { return "x" }}).Parse(src)
	if err != nil { fmt.Println("PARSE ERR:", err); os.Exit(1) }
	data := map[string]interface{}{
		"Name":"T","Email":"t","Balance":1.0,"IsAdmin":false,
		"Packages":[]map[string]interface{}{}, "Transactions":[]map[string]interface{}{},
		"Payments":[]map[string]interface{}{}, "Domains":[]map[string]interface{}{},
	}
	err = t.Execute(os.Stdout, data)
	if err != nil { fmt.Println("\nEXEC ERR:", err) } else { fmt.Println("\nEXEC OK") }
}