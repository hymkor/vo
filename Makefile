NAME=$(lastword $(subst /, ,$(abspath .)))
VERSION=$(shell git.exe describe --tags 2>nul || echo v0.0.0)
GOOPT=-ldflags "-s -w -X main.version=$(VERSION)"
GOEXE=$(shell go env GOEXE)

ifeq ($(OS),Windows_NT)
    SHELL=CMD.EXE
    SET=SET
else
    SET=export
endif

all:
	$(foreach X,$(wildcard internal/*),cd $(X) && go fmt && cd ../.. && ) :
	$(foreach X,$(wildcard cmd/*),cd $(X) && $(SET) "CGO_ENABLED=0" && go build -o ../../$(notdir $(X)$(GOEXE)) $(GOOPT) && cd ../.. && ) :

_package:
	$(MAKE) all
	$(foreach X,$(wildcard cmd/*),zip $(notdir $(X))-$(VERSION)-$(GOOS)-$(GOARCH).zip $(notdir $(X))$(GOEXE) && ) :

package:
	$(SET) "GOOS=windows" && $(SET) "GOARCH=386"   && $(MAKE) _package
	$(SET) "GOOS=windows" && $(SET) "GOARCH=amd64" && $(MAKE) _package

manifest:
	make-scoop-manifest vo*-windows-*.zip > vo.json
	make-scoop-manifest --inline "{ \"description\": \"Show the version number , timestamp and MD5SUM of Windows Executables\" }" showver*-windows-*.zip > showver.json
