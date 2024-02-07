package main

import (
	errname "github.com/Antonboom/errname/pkg/analyzer"
	"github.com/SerjZimmer/devops/cmd/staticlint"
	"github.com/butuzov/mirror"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

func main() {
	var myChecks []*analysis.Analyzer

	myChecks = append(myChecks,
		inspect.Analyzer,
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
	)

	for _, sa := range staticcheck.Analyzers {
		if strings.HasPrefix(sa.Analyzer.Name, "SA") {
			myChecks = append(myChecks, sa.Analyzer)
		}
	}
	for _, sa := range stylecheck.Analyzers {
		if strings.EqualFold(sa.Analyzer.Name, "ST1000") {
			myChecks = append(myChecks, sa.Analyzer)
		}
	}

	myChecks = append(myChecks, errname.New())
	myChecks = append(myChecks, mirror.NewAnalyzer())
	myChecks = append(myChecks, staticlint.Analyzer)

	multichecker.Main(
		myChecks...,
	)
}
