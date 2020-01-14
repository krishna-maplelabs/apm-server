package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/pavius/impi"
)

type consoleErrorReporter struct{}

type stringArrayFlags []string

func (saf *stringArrayFlags) String() string {
    return strings.Join(*saf, ",")
}

func (saf *stringArrayFlags) Set(value string) error {
    *saf = append(*saf, value)
    return nil
}

func (cer *consoleErrorReporter) Report(err impi.VerificationError) {
	fmt.Printf("%s: %s\n", err.FilePath, err.Error())
}

func getVerificationSchemeType(scheme string) (impi.ImportGroupVerificationScheme, error) {
	switch scheme {
	case "stdLocalThirdParty":
		return impi.ImportGroupVerificationSchemeStdLocalThirdParty, nil
	case "stdThirdPartyLocal":
		return impi.ImportGroupVerificationSchemeStdThirdPartyLocal, nil
	default:
		return 0, fmt.Errorf("Unsupported verification scheme: %s", scheme)
	}
}

func run() error {

	var localPrefix = flag.String("local", "", "prefix of the local repository")
	var scheme = flag.String("scheme", "", "verification scheme to enforce. one of stdLocalThirdParty/stdThirdPartyLocal")

	var skipPaths stringArrayFlags
	flag.Var(&skipPaths, "skip", "paths to skip (regex)")

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs)

	// parse flags
	flag.Parse()

	verificationScheme, err := getVerificationSchemeType(*scheme)
	if err != nil {
		return err
	}

	// TODO: can parallelize across root paths
	for argIndex := 0; argIndex < flag.NArg(); argIndex++ {
		rootPath := flag.Arg(argIndex)

		impiInstance, err := impi.NewImpi(numCPUs)
		if err != nil {
			return fmt.Errorf("Failed to create impi: %s", err.Error())
		}

		err = impiInstance.Verify(rootPath, &impi.VerifyOptions{
			SkipTests:     false,
			LocalPrefix:   *localPrefix,
			Scheme:        verificationScheme,
			SkipPaths:     skipPaths,
		}, &consoleErrorReporter{})

		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s PACKAGE [PACKAGE ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	if err := run(); err != nil {
		fmt.Printf("\nimpi verification failed: %s\n", err.Error())
		os.Exit(1)
	}
}
