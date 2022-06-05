module github.com/fuji8/gotypeconverter

go 1.18

require (
	github.com/fatih/structtag v1.2.0
	github.com/gostaticanalysis/analysisutil v0.7.1
	github.com/gostaticanalysis/codegen v0.1.0
	golang.org/x/tools v0.1.9
)

require (
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/gostaticanalysis/comment v1.4.2 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
	golang.org/x/xerrors v0.0.0-20220411194840-2f41105eb62f // indirect
)

replace (
	github.com/gostaticanalysis/codegen => github.com/fuji8/codegen v0.1.1-0.20220501161814-8186a8cd04d8
	golang.org/x/tools => github.com/fuji8/tools v0.1.11-0.20220501160630-8a0bea4d5aba

)
