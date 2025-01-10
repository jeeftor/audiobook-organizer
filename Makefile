build:
	goreleaser build --snapshot --clean --verbose
clean:
	rm -rf ./dist


example/books/afterworldsend:
	@mkdir -p $@

example/books/anythingyoucando:
	@mkdir -p $@

example/books/battleforthestars:
	@mkdir -p $@

download_and_unzip_wget: example/books/afterworldsend example/books/anythingyoucando example/books/battleforthestars
	@wget -P example/books/afterworldsend https://www.archive.org/download/after_worlds_end_2312_librivox/after_worlds_end_2312_librivox_64kb_mp3.zip
	@wget -P example/books/anythingyoucando https://www.archive.org/download/anythingycdo_mn_1302_librivox/anythingycdo_mn_1302_librivox_64kb_mp3.zip
	@wget -P example/books/battleforthestars https://www.archive.org/download/battleforthestars_2401_librivox/battleforthestars_2401_librivox_64kb_mp3.zip
	@unzip -o example/books/afterworldsend/after_worlds_end_2312_librivox_64kb_mp3.zip -d example/books/afterworldsend
	@unzip -o example/books/anythingyoucando/anythingycdo_mn_1302_librivox_64kb_mp3.zip -d example/books/anythingyoucando
	@unzip -o example/books/battleforthestars/battleforthestars_2401_librivox_64kb_mp3.zip -d example/books/battleforthestars
	@rm example/books/afterworldsend/after_worlds_end_2312_librivox_64kb_mp3.zip
	@rm example/books/anythingyoucando/anythingycdo_mn_1302_librivox_64kb_mp3.zip
	@rm example/books/battleforthestars/battleforthestars_2401_librivox_64kb_mp3.zip
	@echo "Files downloaded and unzipped successfully"

download_and_unzip_curl: example/books/afterworldsend example/books/anythingyoucando example/books/battleforthestars
	@curl -L -o example/books/afterworldsend/after_worlds_end_2312_librivox_64kb_mp3.zip https://www.archive.org/download/after_worlds_end_2312_librivox/after_worlds_end_2312_librivox_64kb_mp3.zip
	@curl -L -o example/books/anythingyoucando/anythingycdo_mn_1302_librivox_64kb_mp3.zip https://www.archive.org/download/anythingycdo_mn_1302_librivox/anythingycdo_mn_1302_librivox_64kb_mp3.zip
	@curl -L -o example/books/battleforthestars/battleforthestars_2401_librivox_64kb_mp3.zip https://www.archive.org/download/battleforthestars_2401_librivox/battleforthestars_2401_librivox_64kb_mp3.zip
	@unzip -o example/books/afterworldsend/after_worlds_end_2312_librivox_64kb_mp3.zip -d example/books/afterworldsend
	@unzip -o example/books/anythingyoucando/anythingycdo_mn_1302_librivox_64kb_mp3.zip -d example/books/anythingyoucando
	@unzip -o example/books/battleforthestars/battleforthestars_2401_librivox_64kb_mp3.zip -d example/books/battleforthestars
	@rm example/books/afterworldsend/after_worlds_end_2312_librivox_64kb_mp3.zip
	@rm example/books/anythingyoucando/anythingycdo_mn_1302_librivox_64kb_mp3.zip
	@rm example/books/battleforthestars/battleforthestars_2401_librivox_64kb_mp3.zip
	@echo "Files downloaded and unzipped successfully"