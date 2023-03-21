# durin: Durable Index
Durin is a lightweight, durable key-value store.

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

The bind address can be specified with `-b` and the port with `-p`.

To specify the BasicAuth parameters, set the following environmental variables:
```bash
DURIN_USER=myuser DURIN_PASS=mypass durin
```

To setup the certificates required for https, run `gen-certs.sh`.
```bash
sh gen-certs.sh
```

## HTTP API
```
// Set a key to a value
curl --user <user>:<pass> \
     --cacert durin/certs/tls/ca.crt \
    https://localhost:8045/api/v1/set \
    --data '{"key":<key>,"value":<value>}'

// Get the value of a key
curl --user <user>:<pass> \
     --cacert durin/certs/tls/ca.crt \
     https://localhost:8045/api/v1/get \
     --data '{"key":<key>}'

// Delete a key and its value
curl --user <user>:<pass> \
     --cacert durin/certs/tls/ca.crt \
     https://localhost:8045/api/v1/del \
     --data '{"key":<key>}'
```

## Using TLS
Durin supports TLS encryption via HTTPS. Use the `-c` and `-k` options to specify the TLS certificate and key:
```
# This will generate some sample certificates into the certs/ directory.
sh gen-certs.sh

# Specify the certificate and key like so:
DURIN_USER=frodo DURIN_PASS=baggins durin -p 443 -c durin.crt -k durin.key
```

To use TLS with curl, specify ``--cacert`. A complete example would look like this:
```
git clone https://github.com/chrisvarga/durin && cd durin
make && sudo make install
sh gen-certs.sh
DURIN_USER=frodo DURIN_PASS=baggins durin -d db.json -p 443 -c certs/tls/durin.crt -k certs/tls/durin.key > durin.log 2>&1 &
curl --user frodo:baggins --cacert certs/tls/ca.crt https://localhost/api/v1/set --data '{"key":"foo","value":"bar"}'
curl --user frodo:baggins --cacert certs/tls/ca.crt https://localhost/api/v1/get --data '{"key":"foo"}'
```
