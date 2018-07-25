package manager

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// CommandLineManagerInterface interface
type CommandLineManagerInterface interface {
	Parse(serviceName string) bool
}

// ServiceCommandLineManager structure
type ServiceCommandLineManager struct {
	DebugAddr *string
	GRPCAddr  *string
	ZipkinURL *string
}

// Parse command line argument
func (manager *ServiceCommandLineManager) Parse(serviceName string) bool {
	// Extract command line info
	flag.StringVar(manager.DebugAddr, "debug.addr", ":5060", "Debug and metrics listen address")
	flag.StringVar(manager.GRPCAddr, "grpc.addr", ":5040", "gRPC (HTTP) listen address")

	// Use environment variables, if set. Flags have priority over Env vars.
	if addr := os.Getenv("DEBUG_ADDR"); addr != "" {
		*manager.DebugAddr = addr
	}
	if addr := os.Getenv("GRPC_ADDR"); addr != "" {
		*manager.GRPCAddr = addr
	}
	return true
}

//
func usageFor(fs *flag.FlagSet, short string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "USAGE\n")
		fmt.Fprintf(os.Stderr, "  %s\n", short)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "FLAGS\n")
		w := tabwriter.NewWriter(os.Stderr, 0, 2, 2, ' ', 0)
		fs.VisitAll(func(f *flag.Flag) {
			fmt.Fprintf(w, "\t-%s %s\t%s\n", f.Name, f.DefValue, f.Usage)
		})
		w.Flush()
		fmt.Fprintf(os.Stderr, "\n")
	}
}
