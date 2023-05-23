.PHONY: xml import unizp update  rebuild 
token = ""
all:
# Keep this as the first
all:
	go build -ldflags='-s -w'


xml:
	./nvdbtools cnnvd getxml --token $(token)

import:
	./nvdbtools cnnvd importDB


unzip:
	./nvdbtools cve unzip

update:
	./nvdbtools cve update 

rebuild:
	./nvdbtools cve rebuild

