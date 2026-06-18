package main

import (
	"context"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"time"

	"github.com/a-castellano/home-ip-monitor/app"
	config "github.com/a-castellano/home-ip-monitor/config"
	ipinfo "github.com/a-castellano/home-ip-monitor/ipinfo"
	"github.com/a-castellano/home-ip-monitor/nslookup"
)

// serviceName is the identifier used for syslog messages
const serviceName = "home-ip-monitor"

// main is the entry point of the application
// It initializes logging, configuration, and starts the IP monitoring service
func main() {

	// Configure logger to write to the syslog. You could do this in init(), too.
	logwriter, e := syslog.New(syslog.LOG_INFO, serviceName)
	if e == nil {
		log.SetOutput(logwriter)
		// Remove timestamp since syslog already provides it
		log.SetFlags(0)
	}

	// Now from anywhere else in your program, you can use this:
	log.Print("Loading config")

	// Load application configuration from environment variables
	appConfig, configErr := config.NewConfig()

	if configErr != nil {
		log.Print(configErr.Error())
		os.Exit(1)
	}

	log.Print("Defining http client used by ipInfo package")

	// Create HTTP client with timeout for external API calls
	httpClient := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	// Initialize the IP info requester with the HTTP client
	requester := ipinfo.Realrequester{Client: httpClient}

	log.Print("Defining nslookup client")
	// Initialize DNS lookup client with configured DNS server
	nsLookup := nslookup.DNSLookup{DNSServer: appConfig.DNSServer}
	ctx := context.Background()

	// Start the monitoring process
	if app.Monitor(ctx, requester, nsLookup, appConfig) != nil {
		log.Print("Error running monitor")
		os.Exit(1)
	}

}
