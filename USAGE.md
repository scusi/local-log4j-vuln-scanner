## Usage of log4j-scanner and server

This version of log4j scanner and server (v0.13.1) is meant to run in infrastructures.

The idea is to have a central server running where all the scanners upload their results.
The result files on the central server can then easily analysed for affected machines.

## start a central server

Starting a central log4j collection server is quite easy and straight forward.

Choose the right binary fpr your server from the `dist` directory. 
In the following examples we will run the server on a 64bit linux machine.

Copy `dist/server_linux_amd64/log4j-server` into some directory of your choice 
onto the linux machine that should act as a central server.

For this example we assume you place the server into `/opt/log4j-scanner`.

Change into the directory with the server binary copied in the last step.
Start the server process by executing:
```
log4j-server
```

The above command will start a collection server on all IPs on port 8080 with plain HTTP.
Off course you can change the port and the listening IP address via the `-addr` flag.
You can also turn on SSL,... check out the help, via `-h` flag.

If your server machine has the IP 192.168.1.66, the uploadURL of the server is:
http://192.168.1.66:8080/

## start a scanner

In order to scan a given machine choose the right binary for your operating system and platform.
Bring that binary onto the machine to scan.
Start the scanner.

```
./log4j-scanner -uploadURL http://192.168.1.66:8080/upload /
```

The above command will start a log4j vulnerability scan for all files on disk (starting at the root of the filesystem, "/"), and send the result files to the giben upload URL.
The result file name will have a format as follows:

`<IP>-<HOSTNAME>_<YYYYMMdd_hhmmss>_log4j-vuln-scanner.log`

- <<IP>> will be replaced by the IP of the machine the scanner run on.
- <<HOSTNAME>> will be replaced by the hostname of the machine
- <<YYYYMMdd_hhmmss>> will be replaced by the timestamp when the scan was started

In the default settings the scanner will log and report only found indicators but not a list of all files beeing found and/or scanned.
You can change that behaviour via the `-quiet` flag. When `-quiet=flase` is set on the commandline the scanner will report every processed file.
Please note that turning off the quiet mode will dramatical increase the size of the logfiles. Instead of 4KB a result file will be 400 MB.


