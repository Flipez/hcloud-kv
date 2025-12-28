# Slow key/value store based on Hetzner Cloud Metadata
Implements a simple key/value store by utilizing the firewall resource and its rule descriptions.
As per API spec it supports up to 500 rules with 255 bytes description each.
The data is encoded using ![gob](https://pkg.go.dev/encoding/gob) and compressed using ![zstd](https://pkg.go.dev/github.com/klauspost/compress/zstd).
The actual number of keys and size of values therefore depends on how well they can be compressed.
hetzner-kv supports multiple "databases" provided by a flag.

It uses the token specified via env variable `HCLOUD_TOKEN`

This tool is a PoC only, do not use for anything.

```
NAME:
   hcloud-kv - hetzner cloud key/value store

USAGE:
   hcloud-kv [global options] command [command options] [arguments...]

COMMANDS:
   init, i    initializes a new database
   set, s     sets a key
   get, g     get a value from given key
   list, l    list all keys
   delete, d  delete a given key
   clear, c   delete all keys
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --db value  database to use (default: "0")
   --no-info   Do not print db usage information (default: false)
   --help, -h  show help
```