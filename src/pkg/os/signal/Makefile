# Copyright 2009 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

include ../../../Make.inc

TARG=os/signal
GOFILES=\
	signal.go\
	unix.go\

CLEANFILES+=unix.go

include ../../../Make.pkg

unix.go: ../../syscall/zerrors_$(GOOS)_$(GOARCH).go
	./mkunix.sh $< > $@ || rm -f $@
