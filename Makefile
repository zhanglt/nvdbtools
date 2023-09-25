.PHONY: xml import unizp update  rebuild 
token = $(shell ./gettoken.py )
all:

# Keep this as the first
all:
	go build -ldflags='-s -w'
# @echo  $(token)

xml:
	./nvdbtools cnnvd getxml --token $(token)

import:
	./nvdbtools cnnvd importDB

getcve:
	./getdbfile.sh

unzip:
	./nvdbtools cve unzip

update:
	./nvdbtools cve update 

rebuild:
	./nvdbtools cve rebuild
p1: xml import getcve

p2: unzip update 

prepare: p1 p2 

build: prepare rebuild



