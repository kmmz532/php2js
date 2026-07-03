go build -o php2js.exe ./cmd/php2js/
del /s /q output\src
php2js.exe -input ./pukiwiki-1.5.4_utf8 -output ./output -name pkwk4cf
