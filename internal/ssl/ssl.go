package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
)

type Status int

const (
	StatusOK Status = iota
	StatusWarning
	StatusCritical
	StatusUnknown
)

type CheckerConfig struct {
	ConnectTimeout   time.Duration
	WarnValidity     time.Duration
	CriticalValidity time.Duration
}

// CertStatus represents the status of an SSL certificate
type CertStatus struct {
	Status        Status
	CommonName    string
	Host          string
	Expires       time.Time
	TimeRemaining time.Duration
}

// Check retrieves the SSL certificate chain for a host and returns a map of
// CertStatus structs for each certificate in the chain
func Check(log zerolog.Logger, host string, cfg CheckerConfig) (map[string]CertStatus, error) {
	// remember the checked certs based on their Signature
	checkedCerts := make(map[string]struct{})
	// cert status output
	certStatuses := make(map[string]CertStatus)
	certs, err := lookupCerts(host, cfg.ConnectTimeout)
	if err != nil {
		return certStatuses, err
	}
	// loop over all certs
	// there might be multiple chains, as there may be one or more CAs present on the current system, so we have multiple possible chains
	for _, chain := range certs {
		for _, cert := range chain {
			if _, checked := checkedCerts[string(cert.Signature)]; checked {
				continue
			}
			checkedCerts[string(cert.Signature)] = struct{}{}
			// filter out CA certificates
			if cert.IsCA {
				log.Debug().
					Str("host", host).
					Str("certCommonName", cert.Subject.CommonName).
					Time("expiry", cert.NotAfter).
					Msg(fmt.Sprintf("%-15s - ignoring CA certificate %s", host, cert.Subject.CommonName))
				continue
			}
			var certStatus Status
			remainingValidity := cert.NotAfter.Sub(time.Now())
			if remainingValidity < cfg.CriticalValidity {
				certStatus = StatusCritical
			} else if remainingValidity < cfg.WarnValidity {
				certStatus = StatusWarning
			} else {
				certStatus = StatusOK
			}
			certStatuses[string(cert.Signature)] = CertStatus{
				Status:        certStatus,
				Host:          host,
				CommonName:    cert.Subject.CommonName,
				TimeRemaining: remainingValidity,
				Expires:       cert.NotAfter,
			}
		}
	}
	return certStatuses, nil
}

func lookupCerts(host string, timeout time.Duration) ([][]*x509.Certificate, error) {
	dialer := net.Dialer{Timeout: timeout, Deadline: time.Now().Add(timeout + 5*time.Second)}
	connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:443", host), &tls.Config{ServerName: host})
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	return connection.ConnectionState().VerifiedChains, nil
}
