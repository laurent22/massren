go build
/e/Programmes/7-Zip/7z.exe a -tzip -mx=9 massren.win-386.zip massren.exe
mkdir deploy/releases
mv massren.win-386.zip deploy/releases