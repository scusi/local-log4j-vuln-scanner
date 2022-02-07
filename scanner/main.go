package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"mime/multipart"

	"github.com/hillu/local-log4j-vuln-scanner/filter"
)

var logFile = os.Stdout
var errFile = os.Stderr
var hostname string
var err error
var ip string
var startTime time.Time
var endTime time.Time
var debug bool

func init() {
	hostname, err = os.Hostname()
	if err != nil {
		fmt.Fprintf(errFile, "WARNING: Could not get hostname.")
	}
	startTime = time.Now()
}

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}

// upload file to upload the log to a central server
func upload(filename string) (err error) {
	if uploadURL == "" {
		err = fmt.Errorf("No uploadURL given for upload")
		return err
	}
	if filename == "" {
		err = fmt.Errorf("no filename give for upload")
		return err
	}
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	body := &bytes.Buffer{}
  	writer := multipart.NewWriter(body)
  	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
  	io.Copy(part, file)
  	writer.Close()

  	r, _ := http.NewRequest("POST", uploadURL, body)
  	r.Header.Add("Content-Type", writer.FormDataContentType())
  	client := &http.Client{}
  	res, err := client.Do(r)
	/*
	res, err := http.Post(uploadURL, "binary/octet-stream", file)
	if err != nil {
		return err
	}
	*/
	defer res.Body.Close()
	if debug {
		fmt.Printf("Response StatusCode: %d\n", res.StatusCode)
		message, _ := ioutil.ReadAll(res.Body)
		fmt.Printf(string(message))
	}
	return
}

func handleJar(path string, ra io.ReaderAt, sz int64) {
	if verbose {
		fmt.Fprintf(logFile, "Inspecting %s...\n", path)
	}
	zr, err := zip.NewReader(ra, sz)
	if err != nil {
		fmt.Fprintf(logFile, "cant't open JAR file: %s (size %d): %v\n", path, sz, err)
		return
	}
	for _, file := range zr.File {
		if file.FileInfo().IsDir() {
			continue
		}
		switch strings.ToLower(filepath.Ext(file.Name)) {
		case ".jar", ".war", ".ear":
			fr, err := file.Open()
			if err != nil {
				fmt.Fprintf(logFile, "can't open JAR file member for reading: %s (%s): %v\n", path, file.Name, err)
				continue
			}
			buf, err := ioutil.ReadAll(fr)
			fr.Close()
			if err != nil {
				fmt.Fprintf(logFile, "can't read JAR file member: %s (%s): %v\n", path, file.Name, err)
			}
			handleJar(path+"::"+file.Name, bytes.NewReader(buf), int64(len(buf)))
		default:
			fr, err := file.Open()
			if err != nil {
				fmt.Fprintf(logFile, "can't open JAR file member for reading: %s (%s): %v\n", path, file.Name, err)
				continue
			}

			// Identify class filess by magic bytes
			buf := bytes.NewBuffer(nil)
			if _, err := io.CopyN(buf, fr, 4); err != nil {
				if err != io.EOF && !quiet {
					fmt.Fprintf(logFile, "can't read magic from JAR file member: %s (%s): %v\n", path, file.Name, err)
				}
				fr.Close()
				continue
			} else if !bytes.Equal(buf.Bytes(), []byte{0xca, 0xfe, 0xba, 0xbe}) {
				fr.Close()
				continue
			}
			_, err = io.Copy(buf, fr)
			fr.Close()
			if err != nil {
				fmt.Fprintf(logFile, "can't read JAR file member: %s (%s): %v\n", path, file.Name, err)
				continue
			}
			if info := filter.IsVulnerableClass(buf.Bytes(), file.Name, vulns); info != nil {
				fmt.Fprintf(logFile, "indicator for vulnerable component found in %s (%s): %s %s %s\n",
					path, file.Name, info.Filename, info.Version, info.Vulnerabilities&vulns)
				continue
			}
		}
	}
}

type excludeFlags []string

func (flags *excludeFlags) String() string {
	return fmt.Sprint(*flags)
}

func (flags *excludeFlags) Set(value string) error {
	*flags = append(*flags, filepath.Clean(value))
	return nil
}

func (flags excludeFlags) Has(path string) bool {
	for _, exclude := range flags {
		if path == exclude {
			return true
		}
	}
	return false
}

var excludes excludeFlags
var verbose bool
var logFileName string
var quiet bool
var vulns filter.Vulnerabilities
var ignoreVulns filter.Vulnerabilities = filter.CVE_2021_45046 | filter.CVE_2021_44832
var ignoreV1 bool
var network bool
var uniqLogFileName bool
var uploadURL string

func main() {
	flag.BoolVar(&debug, "debug", false, "prints debug info if set to 'true'")
	flag.Var(&excludes, "exclude", "paths to exclude (can be used multiple times)")
	flag.BoolVar(&verbose, "verbose", false, "log every archive file considered")
	flag.StringVar(&logFileName, "log", "", "log file to write output to")
	flag.BoolVar(&quiet, "quiet", true, "no ouput unless vulnerable")
	flag.BoolVar(&ignoreV1, "ignore-v1", false, "ignore log4j 1.x versions")
	flag.Var(&ignoreVulns, "ignore-vulns", "ignore vulnerabilities")
	flag.BoolVar(&network, "scan-network", false, "search network filesystems")
	flag.BoolVar(&uniqLogFileName, "uniqlogname", true, "auto generate a uniq name for the logfile")
	flag.StringVar(&uploadURL, "uploadURL", "", "URL to upload log file to")

	flag.Parse()

	if len(flag.Args()) <= 1 {
		err = fmt.Errorf("No Path to scan! Please add at least one path to scan.")
		log.Fatal(err)
	}
	if ignoreV1 {
		ignoreVulns |= filter.CVE_2019_17571
	}
	vulns = filter.CheckAllVulnerabilities ^ ignoreVulns

	if !quiet {
		fmt.Printf("%s - a simple local log4j vulnerability scanner\n\n", filepath.Base(os.Args[0]))
	}

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--debug] [--verbose] [--quiet] [--ignore-v1] [--exclude <path>] [--log <file>] [--uploadURL <url>] [--uniqlogname] paths...\n", os.Args[0])
		os.Exit(1)
	}

	// if an uploadURL is given, uniqueLogFileName will be set to `true`.
	if uploadURL != "" {
		uniqLogFileName = true
	}
	if uniqLogFileName {
		ip = GetLocalIP()
		ts := startTime.Format("20060102_150405")
		logFileName = fmt.Sprintf("%s-%s_%s_log4j-vuln-scanner.log", ip, hostname, ts)
	}
	if logFileName != "" {
		f, err := os.Create(logFileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not create log file")
			os.Exit(2)
		}
		logFile = f
		errFile = f
		defer f.Close()
	}

	fmt.Fprintf(logFile, "StartTime of scan: %s\n", startTime)
	fmt.Fprintf(logFile, "Checking for vulnerabilities: %s\n", vulns)

	for _, root := range flag.Args() {
		filepath.Walk(filepath.Clean(root), func(path string, info os.FileInfo, err error) error {
			if isPseudoFS(path) {
				if !quiet {
					fmt.Fprintf(logFile, "Skipping %s: pseudo filesystem\n", path)
				}
				return filepath.SkipDir
			}
			if !network && isNetworkFS(path) {
				if !quiet {
					fmt.Fprintf(logFile, "Skipping %s: network filesystem\n", path)
				}
				return filepath.SkipDir
			}

			if !quiet {
				fmt.Fprintf(logFile, "examining %s\n", path)
			}
			if err != nil {
				fmt.Fprintf(errFile, "%s: %s\n", path, err)
				return nil
			}
			if excludes.Has(path) {
				if !quiet {
					fmt.Fprintf(logFile, "Skipping %s: explicitly excluded\n", path)
				}
				return filepath.SkipDir
			}
			if info.IsDir() {
				return nil
			}
			switch ext := strings.ToLower(filepath.Ext(path)); ext {
			case ".jar", ".war", ".ear":
				f, err := os.Open(path)
				if err != nil {
					fmt.Fprintf(errFile, "can't open %s: %v\n", path, err)
					return nil
				}
				defer f.Close()
				sz, err := f.Seek(0, os.SEEK_END)
				if err != nil {
					fmt.Fprintf(errFile, "can't seek in %s: %v\n", path, err)
					return nil
				}
				handleJar(path, f, sz)
			default:
				return nil
			}
			return nil
		})
	}
	endTime = time.Now()
	fmt.Fprintf(logFile, "EndTime of scan: %s\n", endTime)
	fmt.Fprintf(logFile, "scanning of %s (%s) took %s\n", hostname, ip, endTime.Sub(startTime))
	if !quiet {
		fmt.Println("\nScan finished")
	}
	err := upload(logFileName)
	if err != nil {
		fmt.Println("\nError while uploading logfile")
	} else {
		if !quiet {
			fmt.Printf("scanlog uploaded sucessfully to: %s\n", uploadURL)
		}
	}
}
