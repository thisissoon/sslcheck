package ssl

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog"
)

// Status of SSL certificate expiry
type Status int

const (
	// StatusOK means the cert is not due to expire soon
	StatusOK Status = iota
	// StatusWarning means the cert is due to expire within the
	// configured warning duration
	StatusWarning
	// StatusCritical means the cert is due to expire within the
	// configured critical duration
	StatusCritical
	// StatusUnknown means the cert expirty is unknown
	StatusUnknown
)

// CheckerConfig configures time thresholds for certificate checker
type CheckerConfig struct {
	ConnectTimeout   time.Duration
	WarnValidity     time.Duration
	CriticalValidity time.Duration
}

// CertStatus represents the status of an SSL certificate
type CertStatus struct {
	Status        Status
	CommonName    string
	DNSNames      []string
	Host          string
	NotAfter      time.Time
	TimeRemaining time.Duration
	Issuer        string
}

// Check retrieves the SSL certificate chain for a host and returns a map of
// CertStatus structs for each certificate in the chain
func Check(log zerolog.Logger, host string, cfg CheckerConfig) (*CertStatus, error) {
	// remember the checked certs based on their Signature
	checkedCerts := make(map[string]struct{})
	// cert status output
	var certStatus *CertStatus
	certs, err := lookupCerts(host, cfg.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	for _, cert := range certs {
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
		certStatus = &CertStatus{
			Status:        expiryStatus(cert.NotAfter, cfg.WarnValidity, cfg.CriticalValidity),
			Host:          host,
			CommonName:    cert.Subject.CommonName,
			DNSNames:      cert.DNSNames,
			TimeRemaining: time.Until(cert.NotAfter),
			NotAfter:      cert.NotAfter,
			Issuer:        cert.Issuer.CommonName,
		}
		if err := cert.VerifyHostname(host); err != nil {
			return certStatus, err
		}
	}
	return certStatus, nil
}

func lookupCerts(host string, timeout time.Duration) ([]*x509.Certificate, error) {
	dialer := net.Dialer{Timeout: timeout, Deadline: time.Now().Add(timeout + 5*time.Second)}
	connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:443", host), &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: true, // skip verification as we verify later
	})
	if err != nil {
		return nil, err
	}
	defer connection.Close()
	return connection.ConnectionState().PeerCertificates, nil
}

// expiryStatus returns a status from expiry time and thresholds
func expiryStatus(exp time.Time, warn time.Duration, crit time.Duration) Status {
	r := time.Until(exp)
	if r < crit {
		return StatusCritical
	} else if r < warn {
		return StatusWarning
	}
	return StatusOK
}
