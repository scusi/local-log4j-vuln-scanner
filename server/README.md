# log4j scanner log collection server

## Usage

normally you do not need to do anything more than just start the server executable.

There are several commandline flags available to tweak the behaviour of the server.

To view all available commandline flags use `-h` flag (for help).

```
./log4j-server.exe -h
```

### available flags

```
  -addr string
        http service address (default ":8080")
  -cert string
        pem encoded certificate to use (default "cert.pem")
  -key string
        pem encoded key to use (default "key.pem")
  -showAssets
        if set to 'true' uploadDir will be browsable unter '/files/' (default true)
  -sslOn
        if set to 'true' HTTPS is turned on.
  -templateDir string
        directory with templates, relative to working directory
  -uploadDir string
        directory to upload files to (default "./data")
```

#### addr - listening address

With this flag you can spcify the interface and port the server is listening on.
The default setting does listen on port 8080/TCP on all available interfaces.
On machines with multiple interface this is maybe not what you want. 
In this case use the flag to specify the IP and Port seperated by colon (`:`), as in the following example.

```
-addr 192.168.10.42:8081
```

The above axample will make the server listen only on the given IP (192.168.10.42) and port (8081).

```
-addr :9090
```

The above example will make the server listen on all available interfaces but on port 9090.

#### uploadDir - the directory to upload logfile to

This flag specifies the directory where the uploaded logfiles will be stored.
Make sure you have enough space and the propper permissions, so the server can write to that directory.

```
-uploadDir /path/to/directory
```

#### templateDir - the directory to read custom template from

This flag is usually not needed. If you want to customize the layout and design of the HTML pages shown by the server you can give the server a directory to read template files from.

#### showAssets - make the collected logs browsable, or not

With this flag you can choose to show the collected logfiles and have them browsable.
This is a bool flag, so it can be `true` or `false`.
Default setting is `true`.

If this flag is set to false, the upload directory will not be shown by the webserver and is not browsable.

```
-showAssets=false
```

#### sslOn - turn on TLS/SSL

This flag is a bool flag (can be flase or true).
When true it will turn on TLS for the webserver, default is `false`.

```
-sslOn=true
```

The above flag would turn on TLS.

The next example, which is the default setting, will turn TLS off.

```
-sslOn=false
```

#### cert - TLS certificate to use

When you turn on TLS you also need to have a TLS certificate and the corresponding key, see next section.
The certificate must be PEM encoded.

```
-cert /path/to/cert.pem
```

#### key - TLS key to use

When you turn on TLS, you also need to have a TLS key that fits to the TLS certificate (see last section).
The key must be PEM encoded and must have no passphrase set.

```
-key /path/to/key.pem
```


