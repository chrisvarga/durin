# durin: Durable Index
Durin is a lightweight, durable key-value store with a Redis-like syntax.

## Usage
To build and install:
```bash
make && make install
```
> **Note:** `make install` installs to /usr/local/bin by default and may require `sudo`

By default, durin runs in ephemeral mode, where keys are only stored in memory.
To run in durable mode, specify the `-d` option with the storage path:
```bash
durin -d /path/to/database
```
Durin will then write the database to disk once every second if there are any
data changes. The data is stored in json format. With the `-d` option, durin
also will read the database file into memory on startup if it is valid json.

## Syntax
```
set  <key> <value>  // set a key to a value
get  <key>          // get the value of a key
del  <key>          // delete a key and its value
keys [prefix]       // display keys starting with prefix
```

## Examples
```bash
$ nc localhost 8045
keys
[]
set foo bar
OK
keys
["foo"]
get foo
bar
set spam eggs
OK
keys
["foo","spam"]
keys f
["foo"]
```
