package flags

import (
	"log"
	"strconv"
	"strings"
)

func ParseIdAndPort(serviceAndPort string) (string, int) {
	parts := strings.Split(serviceAndPort, ":")

	if len(parts) != 2 {
		log.Fatalf("Fake service '%s' is in invalid format, example 'foo-service:8081'", serviceAndPort)
	}

	serviceID := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf("Fake service '%s' is in invalid format, example 'foo-service:8081'", serviceAndPort)
	}

	return serviceID, port
}
