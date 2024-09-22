# Concurrent File Downloader

A command-line tool written in Go for downloading multiple files concurrently, complete with progress bars and customizable options.

# Build

### 1. Clone the repo

```sh
git clone https://github.com/yasinokat/concurrent-downloader.git
cd concurrent-downloader
```

### 2. Get dependencies

```sh
go mod tidy
```

### 3. Build the app

```sh
go build -o downloader
```

### 4. Show help

```sh
./downloader -h
```

## Example

```sh
 ./downloader -o downloads -n 3     http://ipv4.download.thinkbroadband.com/512MB.zip     http://ipv4.download.thinkbroadband.com/1GB.zip     http://proof.ovh.net/files/1Gb.dat
```

This will start 3 downloads concurrently
