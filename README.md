## Download batch tool

### Style 1: Download URLs by URL template

Example command:

```bash
go run main.go -url="https://www.example.com/images/{%03d}.png" -from=1 -to=5 -outputDir="/tmp/download"
```

This command will download these files:

```
https://www.example.com/images/001.png
https://www.example.com/images/002.png
https://www.example.com/images/003.png
https://www.example.com/images/004.png
https://www.example.com/images/005.png
```

Flags description:
- `url`: URL template with *leading zero pattern* which is openned by `{` and closed by `}` character.
- `from`: Start number.
- `to`: End number.
- `outputDir`: The path of output directory which contains the downloaded files.

### Style 2: Download URLs from file

You must have a **text** file contains URLs line by line.

Example command:

```bash
go run main.go -file="/tmp/links.txt" -outputDir="/tmp/download"
```

Flags description:
- `file`: The path of text file.
- `outputDir`: The path of output directory which contains the downloaded files.

### Batch size

Default batch size is **4** - download 4 files simultaneously. You can pass batch size value via flag `batchSize`.

For example:

```bash
go run main.go -url="https://www.example.com/images/{%03d}.png" -from=1 -to=5 -outputDir="/tmp/download" -batchSize=10
```
