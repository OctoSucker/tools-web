module github.com/OctoSucker/skill-web

go 1.24.4

toolchain go1.24.10

require (
	github.com/OctoSucker/octosucker-skill v0.0.0
	github.com/go-rod/rod v0.116.2
	golang.org/x/net v0.34.0
)

require (
	github.com/ysmood/fetchup v0.2.3 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.40.0 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.9.0 // indirect
)

replace github.com/OctoSucker/octosucker-skill => ../octosucker-skill
