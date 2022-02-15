# Simple local log4j vulnerability scanner

This is a fork from [github.com/hillu/local-log4j-vuln-scanner](https://github.com/hillu/local-log4j-vuln-scanner).
Compared to the original Version from hillu this fork is modified to better scope with large infrastructures.

The main differences are:
- scanner supports uploading logfiles of scans to a central collection server.
- collection server to receive and hold the logfiles, see [server/](server/)
- containes a version flag `scanner -version` will show you version information of the binary
- logfiles are auto named with HOSTNAME, IP-ADDRESS, and a timestamp when the scan took place
- the `-quiet` flag is set to true by default
- `-uploadURL` flag can be used to set a URL scanlogs will be uploaded to
- scanlog contains a timestamp when the scan started and ended and how long it took.


![logo](logo.png)

(Written in Go because, you know, "write once, run anywhere.")

This is a simple tool that can be used to find vulnerable instances of
log4j 1.x and 2.x in installations of Java software such as web
applications. JAR and WAR archives are inspected and class files that
are known to be vulnerable are flagged. The scan happens recursively:
WAR files containing WAR files containing JAR files containing
vulnerable class files ought to be flagged properly.

Currently recognized vulnerabilities are:
- CVE-2019-17571 (1.x)
- CVE-2021-44228
- CVE-2021-45105
- CVE-2021-45046 (not reported by default due to lower severity)
- CVE-2021-44832 (not reported by default due to lower severity)

The scan tool currently checks for known build artifacts that have
been obtained through Maven. From-source rebuilds as they are done for
Linux distributions may not be recognized.

Also included is a simple patch tool that can be used to patch out bad
classes from JAR files by rewriting the ZIP archive structure.

Binaries for x86_64 Windows, Linux, FreeBSD and MacOSX for tagged releases are
provided via the
[Releases](https://github.com/scusi/local-log4j-vuln-scanner/releases)
page.

# Using the scanner

```
$ ./local-log4j-vuln-scanner [-version] [-verbose] [-quiet] \
    [-ignore-v1] [-ignore-vulns=...] \
    [-exclude /path/to/exclude …] \
	[-scan-network] \
	[-log /path/to/file.log] \
	[-uniqlogname] \
	[-uploadURL http://server:88/upload] \
    /path/to/app1 /path/to/app2 …
```
The `-version`flag will show version information and exit.
```
Version: 0.13.3, branch: master, commit: 843c6f0260c2cd9c9762877ed1f6aff00265cf5c
```

The `-verbose` flag will show every .jar and .war file checked, even if no problem is found.

The `-quiet` flag will supress output except for indicators of a known vulnerability.

The `-ignore-v1` flag will _exclude_ checks for log4j 1.x vulnerabilities.

The `-ignore-vulns` flag allows _excluding_ checks for specific
vulnerabilities. e.g. `-ignore-vulns=CVE-2021-45046,CVE-2021-44832`.
To check for all known vulnerabilities, pass an empty list like so:
`-ignore-vulns=`

The `-log` flag allows everythig to be written to a log file instead of stdout/stderr.

Use the `-exclude` flag to exclude subdirectories from being scanned. Can be used multiple times.

The `-scan-network` flag tells the scanner to search network filesystems (disabled by default). This has not been implemented for Windows.

The `-uniqlogname` flag tells the scanner to generate a host uniq log file name in the format `$IP-$HOSTNAME_log4j-scanner.log`
This option includes `--log`
`$IP` will be the local IP address, non loopback.
`$HOSTNAME` will be the configured hostname.

The `-uploadURL` flag tells the scanner to upload the logfile to the given URL.
This option implies `--uniqlogname`.

If class files indicating one of the vulnerabilities are found,
messages like the following are printed to standard output:
``` console
./local-log4j-vuln-scanner - a simple local log4j vulnerability scanner

Checking for vulnerabilities: CVE-2019-17571, CVE-2021-44228, CVE-2021-45105
examining /path/to/vuln/log4shell-vulnerable-app-0.0.1-SNAPSHOT.war
indicator for vulnerable component found in /path/to/vuln/Downloads/log4shell-vulnerable-app-0.0.1-SNAPSHOT.war::WEB-INF/lib/log4j-core-2.14.1.jar (org/apache/logging/log4j/core/net/JndiManager.class): JndiManager.class log4j 2.14.0-2.14.1 CVE-2021-44228, CVE-2021-45105

Scan finished
```

# Using the patch tool

**Caution:** Use this at your own risk and keep the original JAR files.
```
$ ./local-log4j-vuln-patcher log4j-core-2.14.1.jar log4j-core-2.14.1-patched.jar
Filtering out org/apache/logging/log4j/core/pattern/MessagePatternConverter.class (log4j 2.14)
Filtering out org/apache/logging/log4j/core/net/JndiManager.class (log4j 2.14.0-2.14.1)

Writing to log4j-core-2.14.1-patched.jar done
```

# Building from source

## building with goreleaser

Install [goreleaser](https://goreleaser.com/)

and execute `goreleaser build` from the source directory.

A [.goreleaser.yaml](.goreleaser) config file for [goreleaser](https://goreleaser.com/) is included in the repository.

## manual building
Install a [Go compiler](https://golang.org/dl).

Run the following commands in the checked-out repository.
You also have to replace {{.Version}}, {{.Branch}} and {{.Commit}} with the approriate values, like `v0.0.1`, `master`, `02c0febabe`.

```
go build -ldflags="-s -w -X main.version={{.Version}} -X main.branch={{.Branch}} -X main.commit={{.Commit}}" -o local-log4j-vuln-scanner ./scanner
go build -ldflags="-s -w -X main.version={{.Version}} -X main.branch={{.Branch}} -X main.commit={{.Commit}}" -o local-log4j-vuln-patcher ./patcher
go build -ldflags="-s -w -X main.version={{.Version}} -X main.branch={{.Branch}} -X main.commit={{.Commit}}" -o local-log4j-vuln-server ./server
```
(Add the appropriate `.exe` extension on Windows systems, of course.)

# License

GNU General Public License, version 3

# Author

Hilko Bengen <<bengen@hilluzination.de>>

the following feature where added by Florian Walther <<flw@scu.si>>:
- uploading logfiles to a central server
- time meassure for scanning
- auto generated logfile name, to have hostname and ip in the log file name
- a central log collection server
- added version information
- goreleaser support

