include $(GOROOT)/src/Make.inc

TARG=remyoudompheng/xz
CGOFILES=reader.go\
	 writer.go

GOFILES=enums.go

include $(GOROOT)/src/Make.pkg

examples:
	gomake -C examples

format:
	gofmt -l -s -w *.go
