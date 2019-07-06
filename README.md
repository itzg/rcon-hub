Provides an SSH server that enables rcon connections to configured game servers

## Configuration

Create a YAML file named `config.yml` and place it in the current directory, `/etc/rcon-hub`, or 
in a `.rcon-hub` directory located in your home directory.

Top-level keys of the configuration include:

- `host-key-file` : PEM encoded file that will be used for the SSH host key file. If it doesn't exist, an RSA
  private key file will be created. If not specified, a new one is generated in-memory at startup; however, this
  will require you remove the known hosts entry on each startup
- `history-size` : Size of command history. Default is 100.
- `bind` : the `host:port` where SSH connections are accepted. Default is `:2222`.

### Users

One or more users can be declared under the `users` key with a key-object entry each. 
The sub-key is the username and object requires:
- `password` : the SSH password of the user 

### Connections

One or more RCON connections can be declared under the `connections` key with a key-object entry each.
The sub-key is the name of the connection that will be used with the hub's `connect` command. The object requires:
- `address` : a `host:port` declaration of the RCON endpoint
- `password` : the RCON password to authenticate with the endpoint

### Example

The following example declares one SSH user with the username `testing` and a password of `pw`. 
It also declares a single RCON connection named `local-mc` that will connect to `localhost:25575` with the
password `minecraft`.

```yaml
users:
  testing:
    password: pw
connections:
  local-mc:
    address: localhost:25575
    password: minecraft
host-key-file: host_key.pem
```

## Usage

With the configuration described above, start the hub:

```
rcon-hub
```

Passing `help` or `--help` will display additional command-line options.

## Docker Usage

A containerized version is provided at `itzg/rcon-hub`. It exposes the SSH port at 2222 and declares two volumes:
- `/etc/rcon-hub` : this is where the `config.yml` file is loaded
- `/data` : by default the container will write the host key file to `/data/host_key.pem`, but a pre-generated
  key file can be mounted at that path
  
Rather than using the config file, the following environment variables are supported:
- `RH_USER` : the username to register for SSH authentication. Default is `user`.
- `RH_PASSWORD` : if specified, a user with username `$RH_USER` will be registered with the given password 
- `RH_CONNECTION` : a space delimited list of connection definitions of the form `name=password@host:port`

## Connecting

You can connect to the running hub server using any ssh client and using password authentication, such as:

```
ssh -p 2222 testing@localhost
```

## Shell usage

Once connected, you interact with a simple shell environment. The `help` command lists the other commands
available within the shell. The command history can be accessed by pressing the up/down arrow keys.