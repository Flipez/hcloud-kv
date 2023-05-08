# key/value store based on Hetzner Cloud Labels
Implements a simple key/value store by utilizing the ssh-key resource and its labels.
As per API spec it supports keys and values with a max lenght of 63 chars.
hetzner-kv supports multiple "databases" provided by a flag.

It uses the token specified via env variable `HCLOUD_TOKEN`

This tool is a PoC only, do not use for anything.

```
NAME:
   hcloud-kv - hetzner cloud key/value store

USAGE:
   hcloud-kv [global options] command [command options] [arguments...]

COMMANDS:
   init, i  initializes a new database
   set, s   sets a key
   get, g   get a value from given key
   list, l  list all keys
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --db value  database to use (default: "0")
   --help, -h  show help
```